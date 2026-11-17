package users

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrNotFound    = errors.New("user not found")
	ErrEmailExists = errors.New("email already registered")
)

type User struct {
	ID           string    `json:"id" bson:"_id"`
	Email        string    `json:"email" bson:"email"`
	DisplayName  string    `json:"displayName" bson:"displayName"`
	PasswordHash string    `json:"-" bson:"passwordHash"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
}

type Repository interface {
	Create(ctx context.Context, user User) error
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, userID string) (User, error)
	UpdateProfile(ctx context.Context, userID string, displayName string) (User, error)
}

type MemoryRepository struct {
	mu      sync.RWMutex
	users   map[string]User
	byEmail map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:   make(map[string]User),
		byEmail: make(map[string]string),
	}
}

func (r *MemoryRepository) Create(_ context.Context, user User) error {
	if user.ID == "" || user.Email == "" {
		return errors.New("invalid user payload")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byEmail[user.Email]; exists {
		return ErrEmailExists
	}
	if _, exists := r.users[user.ID]; exists {
		return errors.New("user ID already exists")
	}

	r.users[user.ID] = user
	r.byEmail[user.Email] = user.ID
	return nil
}

func (r *MemoryRepository) GetByEmail(_ context.Context, email string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[email]
	if !ok {
		return User{}, ErrNotFound
	}
	user, ok := r.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (r *MemoryRepository) GetByID(_ context.Context, userID string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[userID]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (r *MemoryRepository) UpdateProfile(_ context.Context, userID string, displayName string) (User, error) {
	if displayName == "" {
		return User{}, errors.New("display name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[userID]
	if !ok {
		return User{}, ErrNotFound
	}
	user.DisplayName = displayName
	user.UpdatedAt = time.Now()
	r.users[userID] = user
	return user, nil
}

func NextUserID() string {
	return primitive.NewObjectID().Hex()
}
