package store

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"yearplanner/internal/auth"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	emailQueueClaimDuration = 30 * time.Second
	emailQueueMaxAttempts   = 5
)

type MongoPasswordResetEmailQueue struct {
	collection *mongo.Collection
	key        [32]byte
}

type passwordResetEmailJob struct {
	ID             primitive.ObjectID `bson:"_id"`
	Recipient      string             `bson:"recipient"`
	EncryptedLink  []byte             `bson:"encryptedLink"`
	Attempts       int                `bson:"attempts"`
	AvailableAt    time.Time          `bson:"availableAt"`
	ClaimExpiresAt time.Time          `bson:"claimExpiresAt"`
}

func NewMongoPasswordResetEmailQueue(collection *mongo.Collection, encryptionSecret string) *MongoPasswordResetEmailQueue {
	return &MongoPasswordResetEmailQueue{
		collection: collection,
		key:        sha256.Sum256([]byte(encryptionSecret)),
	}
}

func (q *MongoPasswordResetEmailQueue) EnsureIndexes(ctx context.Context) error {
	_, err := q.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "availableAt", Value: 1}, {Key: "claimExpiresAt", Value: 1}},
		Options: options.Index().SetName("password_reset_email_queue_ready"),
	})
	return err
}

func (q *MongoPasswordResetEmailQueue) Enqueue(ctx context.Context, email string, link string) error {
	encryptedLink, err := q.encrypt(email, link)
	if err != nil {
		return fmt.Errorf("encrypt password reset email: %w", err)
	}
	now := time.Now().UTC()
	_, err = q.collection.InsertOne(ctx, bson.D{
		{Key: "recipient", Value: email},
		{Key: "encryptedLink", Value: encryptedLink},
		{Key: "attempts", Value: 0},
		{Key: "availableAt", Value: now},
		{Key: "createdAt", Value: now},
	})
	if err != nil {
		return fmt.Errorf("queue password reset email: %w", err)
	}
	return nil
}

func (q *MongoPasswordResetEmailQueue) Run(ctx context.Context, sender auth.PasswordResetEmailSender) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		processed, err := q.processOne(ctx, sender)
		if err != nil && ctx.Err() == nil {
			slog.Error("password reset email queue processing failed", "error", err)
		}
		if processed && err == nil {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (q *MongoPasswordResetEmailQueue) processOne(ctx context.Context, sender auth.PasswordResetEmailSender) (bool, error) {
	now := time.Now().UTC()
	filter := bson.D{
		{Key: "availableAt", Value: bson.D{{Key: "$lte", Value: now}}},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "claimExpiresAt", Value: bson.D{{Key: "$exists", Value: false}}}},
			bson.D{{Key: "claimExpiresAt", Value: bson.D{{Key: "$lte", Value: now}}}},
		}},
	}
	var job passwordResetEmailJob
	err := q.collection.FindOneAndUpdate(
		ctx,
		filter,
		bson.D{
			{Key: "$set", Value: bson.D{{Key: "claimExpiresAt", Value: now.Add(emailQueueClaimDuration)}}},
			{Key: "$inc", Value: bson.D{{Key: "attempts", Value: 1}}},
		},
		options.FindOneAndUpdate().SetSort(bson.D{{Key: "availableAt", Value: 1}}).SetReturnDocument(options.After),
	).Decode(&job)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claim password reset email: %w", err)
	}

	link, err := q.decrypt(job.Recipient, job.EncryptedLink)
	if err != nil {
		if _, deleteErr := q.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: job.ID}}); deleteErr != nil {
			return true, fmt.Errorf("delete unreadable password reset email: %w", deleteErr)
		}
		slog.Error("discarded unreadable password reset email")
		return true, nil
	}

	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	sendErr := sender.SendPasswordReset(sendCtx, job.Recipient, link)
	cancel()
	if sendErr == nil {
		_, err = q.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: job.ID}})
		if err != nil {
			return true, fmt.Errorf("delete sent password reset email: %w", err)
		}
		return true, nil
	}

	if job.Attempts >= emailQueueMaxAttempts {
		_, err = q.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: job.ID}})
		if err != nil {
			return true, fmt.Errorf("delete failed password reset email: %w", err)
		}
		slog.Error("password reset email delivery exhausted retries", "error", sendErr)
		return true, nil
	}

	delay := time.Minute * time.Duration(1<<(job.Attempts-1))
	_, err = q.collection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: job.ID}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "availableAt", Value: time.Now().UTC().Add(delay)},
			{Key: "claimExpiresAt", Value: time.Time{}},
		}}},
	)
	if err != nil {
		return true, fmt.Errorf("schedule password reset email retry: %w", err)
	}
	slog.Error("password reset email delivery failed; retry scheduled", "attempt", job.Attempts, "error", sendErr)
	return true, nil
}

func (q *MongoPasswordResetEmailQueue) encrypt(recipient string, link string) ([]byte, error) {
	block, err := aes.NewCipher(q.key[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	payload := aead.Seal(nonce, nonce, []byte(link), []byte(recipient))
	return payload, nil
}

func (q *MongoPasswordResetEmailQueue) decrypt(recipient string, payload []byte) (string, error) {
	block, err := aes.NewCipher(q.key[:])
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < aead.NonceSize() {
		return "", errors.New("invalid encrypted email payload")
	}
	nonce, ciphertext := payload[:aead.NonceSize()], payload[aead.NonceSize():]
	link, err := aead.Open(nil, nonce, ciphertext, []byte(recipient))
	if err != nil {
		return "", fmt.Errorf("decrypt password reset email: %w", err)
	}
	return string(link), nil
}

var _ auth.PasswordResetEmailQueue = (*MongoPasswordResetEmailQueue)(nil)
