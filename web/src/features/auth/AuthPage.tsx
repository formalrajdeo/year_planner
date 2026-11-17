import { useState, type FormEvent } from "react";
import {
  ArrowRight,
  CalendarDays,
  Compass,
  Eye,
  EyeOff,
  LockKeyhole,
  Mail,
  Sparkles,
  X,
} from "lucide-react";
import type { AuthCredentials } from "../../lib/api";

export type AuthMode = "login" | "register";

const currentYear = new Date().getFullYear();
const decorativeWeeks = Array.from({ length: 52 }, (_, index) => index);
const months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];

export function AuthPage({
  onSubmit,
  error,
}: {
  onSubmit: (mode: AuthMode, credentials: AuthCredentials) => Promise<void>;
  error: string;
}) {
  const [mode, setMode] = useState<AuthMode>("login");
  const [displayName, setDisplayName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    try {
      await onSubmit(mode, {
        email: email.trim(),
        password,
        ...(mode === "register" ? { displayName: displayName.trim() } : {}),
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="auth-page min-h-screen">
      <section className="auth-story">
        <div className="story-noise" aria-hidden="true" />
        <header className="brand brand-light">
          <span className="brand-mark">
            <CalendarDays size={19} strokeWidth={2.2} />
          </span>
          <span>Year Planner</span>
        </header>

        <div className="story-copy">
          <p className="eyebrow eyebrow-light">
            <Sparkles size={14} />
            Make the year feel yours
          </p>
          <h1>
            A little more
            <br />
            <span>intention,</span>
            <br />
            all year long.
          </h1>
          <p className="story-description">
            Bring your plans, habits, and big ideas into one calm place. See
            the whole year, then make today count.
          </p>
        </div>

        <YearGridIllustration />

        <div className="story-footer">
          <span>One year. Weeks of small chances to begin again.</span>
          <span className="footer-sparkle">
            <Sparkles size={15} />
          </span>
        </div>
      </section>

      <section className="auth-panel">
        <div className="auth-panel-top">
          <a className="brand brand-dark" href="/" aria-label="Year Planner home">
            <span className="brand-mark">
              <CalendarDays size={18} strokeWidth={2.2} />
            </span>
            <span>Year Planner</span>
          </a>
          <span className="auth-panel-note">
            <LockKeyhole size={14} />
            PRIVATE BY DESIGN
          </span>
        </div>

        <div className="auth-form-wrap">
          <div className="auth-heading">
            <p className="eyebrow eyebrow-green">
              <span className="eyebrow-dot" />
              A fresh page awaits
            </p>
            <h2>{mode === "login" ? "Welcome back" : "Start with a year"}</h2>
            <p>
              {mode === "login"
                ? "Sign in to pick up where you left off."
                : "Create your account and make space for what matters."}
            </p>
          </div>

          <div className="auth-tabs" role="tablist" aria-label="Account options">
            <button
              type="button"
              role="tab"
              aria-selected={mode === "login"}
              className={mode === "login" ? "auth-tab active" : "auth-tab"}
              onClick={() => setMode("login")}
            >
              Sign in
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={mode === "register"}
              className={mode === "register" ? "auth-tab active" : "auth-tab"}
              onClick={() => setMode("register")}
            >
              Create account
            </button>
          </div>

          <form className="auth-form" onSubmit={submit}>
            {mode === "register" && (
              <label className="field">
                <span>Your name</span>
                <div className="input-shell">
                  <Compass size={17} aria-hidden="true" />
                  <input
                    autoComplete="name"
                    name="displayName"
                    value={displayName}
                    onChange={(event) => setDisplayName(event.target.value)}
                    placeholder="How should we address you?"
                    required
                  />
                </div>
              </label>
            )}

            <label className="field">
              <span>Email address</span>
              <div className="input-shell">
                <Mail size={17} aria-hidden="true" />
                <input
                  autoComplete="email"
                  name="email"
                  type="email"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  placeholder="you@example.com"
                  required
                />
              </div>
            </label>

            <label className="field">
              <span>Password</span>
              <div className="input-shell">
                <LockKeyhole size={17} aria-hidden="true" />
                <input
                  autoComplete={mode === "login" ? "current-password" : "new-password"}
                  minLength={8}
                  name="password"
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder="At least 8 characters"
                  required
                />
                <button
                  className="password-toggle"
                  type="button"
                  aria-label={showPassword ? "Hide password" : "Show password"}
                  onClick={() => setShowPassword((visible) => !visible)}
                >
                  {showPassword ? <EyeOff size={17} /> : <Eye size={17} />}
                </button>
              </div>
            </label>

            {mode === "login" && (
              <div className="form-meta">
                <span className="session-note">Your sign-in stays active for this session.</span>
                <span className="forgot-link" aria-disabled="true">
                  Recovery coming soon
                </span>
              </div>
            )}

            {error && (
              <div className="form-error" role="alert">
                <X size={16} />
                <span>{error}</span>
              </div>
            )}

            <button className="submit-button" type="submit" disabled={isSubmitting}>
              <span>
                {isSubmitting
                  ? "One moment..."
                  : mode === "login"
                    ? "Sign in to your planner"
                    : "Create your account"}
              </span>
              <ArrowRight size={17} />
            </button>
          </form>

          <div className="security-note">
            <LockKeyhole size={14} />
            <span>Your plans are personal and private by default.</span>
          </div>
        </div>

        <footer className="auth-panel-footer">
          <span>© {currentYear} Year Planner</span>
          <span>Small steps add up.</span>
        </footer>
      </section>
    </main>
  );
}

function YearGridIllustration() {
  return (
    <div className="year-illustration" aria-label="Illustration of the year at a glance">
      <div className="illustration-top">
        <div>
          <p className="illustration-label">YOUR YEAR, IN VIEW</p>
          <p className="illustration-year">{currentYear}</p>
        </div>
        <span className="illustration-badge">
          <span />
          A year of possibility
        </span>
      </div>
      <div className="illustration-months" aria-hidden="true">
        {months.map((month) => (
          <span key={month}>{month}</span>
        ))}
      </div>
      <div className="week-illustration-grid" aria-hidden="true">
        {decorativeWeeks.map((week) => (
          <span
            className={`week-illustration-cell ${week % 9 === 2 || week % 13 === 4 ? "soft" : ""} ${week === 45 ? "current" : ""}`}
            key={week}
          />
        ))}
      </div>
      <div className="illustration-bottom">
        <span>
          <span className="legend-dot filled" />
          Illustrative weeks
        </span>
        <span>
          <span className="legend-dot current" />
          Highlighted week
        </span>
      </div>
      <p className="illustration-caption">
        A visual preview · planning data arrives in a later phase
      </p>
    </div>
  );
}
