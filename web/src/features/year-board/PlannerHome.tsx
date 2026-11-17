import { useState, type ReactNode } from "react";
import {
  ArrowDownRight,
  ArrowUpRight,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Compass,
  Leaf,
  Sparkles,
  Target,
} from "lucide-react";
import type { PublicUser } from "../../lib/api";

const currentYear = new Date().getFullYear();

export function PlannerHome({
  user,
  profileLoading,
  profileError,
  logoutError,
  isSigningOut,
  onLogout,
}: {
  user: PublicUser;
  profileLoading: boolean;
  profileError: boolean;
  logoutError: string;
  isSigningOut: boolean;
  onLogout: () => void;
}) {
  const [year, setYear] = useState(currentYear);
  const name = user.displayName.trim().split(/\s+/)[0] || "there";

  return (
    <div className="planner-app">
      <aside className="planner-sidebar">
        <a className="brand brand-dark sidebar-brand" href="/" aria-label="Year Planner home">
          <span className="brand-mark">
            <CalendarDays size={18} strokeWidth={2.2} />
          </span>
          <span>Year Planner</span>
        </a>

        <div className="sidebar-year-label">YOUR SPACE</div>
        <div className="workspace-switcher">
          <span className="workspace-avatar">{name.charAt(0).toUpperCase()}</span>
          <span className="workspace-info">
            <strong>Personal planner</strong>
            <small>{user.email}</small>
          </span>
        </div>

        <nav className="planner-nav" aria-label="Planner navigation">
          <a className="nav-item active" href="#year-board" aria-current="page">
            <CalendarDays size={18} />
            <span>Year board</span>
            <span className="nav-indicator" />
          </a>
          <span className="nav-item nav-item-disabled" aria-disabled="true">
            <Leaf size={18} />
            <span>Habits</span>
            <small>SOON</small>
          </span>
          <span className="nav-item nav-item-disabled" aria-disabled="true">
            <Target size={18} />
            <span>Quarterly goals</span>
            <small>SOON</small>
          </span>
        </nav>

        <div className="sidebar-bottom">
          <div className="sidebar-note">
            <span className="note-icon">
              <Sparkles size={16} />
            </span>
            <p>Make room for progress, not perfection.</p>
          </div>
          <div className="sidebar-user">
            <span className="user-avatar">{name.charAt(0).toUpperCase()}</span>
            <span className="user-details">
              <strong>{user.displayName}</strong>
              <small>Personal account</small>
            </span>
            <button
              className="logout-button"
              type="button"
              onClick={onLogout}
              disabled={isSigningOut}
              aria-label="Sign out"
              title="Sign out"
            >
              <ArrowUpRight size={17} />
            </button>
          </div>
        </div>
      </aside>

      <main className="planner-main" id="year-board">
        <header className="planner-topbar">
          <div className="breadcrumb">
            <span>Personal</span>
            <ChevronRight size={14} />
            <strong>Year board</strong>
          </div>
          <div className="topbar-actions">
            <span className="topbar-date">
              {new Intl.DateTimeFormat("en", {
                weekday: "short",
                month: "short",
                day: "numeric",
              }).format(new Date())}
            </span>
            <button
              className="topbar-avatar"
              type="button"
              onClick={onLogout}
              title="Sign out"
              aria-label="Sign out"
              disabled={isSigningOut}
            >
              {name.charAt(0).toUpperCase()}
            </button>
          </div>
        </header>

        <div className="planner-content">
          {profileError && (
            <div className="notice notice-warning" role="status">
              Profile refresh failed. Showing the account details from sign-in.
            </div>
          )}
          {logoutError && (
            <div className="notice notice-error" role="alert">
              {logoutError}
            </div>
          )}

          <section className="welcome-row">
            <div>
              <p className="eyebrow eyebrow-green">
                <span className="eyebrow-dot" />
                YOUR YEAR, IN VIEW
              </p>
              <h1>Good to see you, {name}.</h1>
              <p className="welcome-copy">
                Your planning space is ready. Let’s make this year count.
              </p>
            </div>
            <div className="year-picker" aria-label="Select year">
              <button
                type="button"
                aria-label="Previous year"
                onClick={() => setYear((value) => value - 1)}
              >
                <ChevronLeft size={17} />
              </button>
              <span>{year}</span>
              <button
                type="button"
                aria-label="Next year"
                onClick={() => setYear((value) => value + 1)}
              >
                <ChevronRight size={17} />
              </button>
            </div>
          </section>

          <section className="year-overview-card">
            <div className="overview-copy">
              <div className="overview-heading">
                <span className="overview-icon">
                  <Compass size={20} />
                </span>
                <span className="overview-kicker">A YEAR WITH ROOM TO GROW</span>
              </div>
              <h2>{year} is yours to shape.</h2>
              <p>
                Your annual plan will bring weekly focus, daily habits, and
                meaningful goals together in one view.
              </p>
              <span className="preview-tag">
                <Sparkles size={13} />
                Planner preview
              </span>
            </div>
            <div className="overview-art" aria-hidden="true">
              <div className="art-orbit orbit-one" />
              <div className="art-orbit orbit-two" />
              <div className="art-center">
                <CalendarDays size={29} />
              </div>
              <span className="art-star star-one">✳</span>
              <span className="art-star star-two">✦</span>
              <span className="art-leaf">
                <Leaf size={22} />
              </span>
            </div>
          </section>

          <section className="board-heading">
            <div>
              <h2>Your year board</h2>
              <p>A clear view of the weeks ahead.</p>
            </div>
            <span className="board-status">
              <span />
              Ready for your first plan
            </span>
          </section>

          {profileLoading && <div className="profile-loading">Refreshing your profile…</div>}

          <section className="empty-board">
            <div className="empty-board-mark">
              <CalendarDays size={23} />
              <span>
                <Sparkles size={13} />
              </span>
            </div>
            <p className="empty-kicker">A FRESH PAGE</p>
            <h3>Your year is waiting to take shape.</h3>
            <p className="empty-copy">
              Annual plans and weekly planning are coming next. Your account is
              connected and ready for the planning features.
            </p>
            <div className="empty-progress" aria-label="Planning board not yet configured">
              <span />
              <span />
              <span />
              <span />
            </div>
            <span className="empty-progress-caption">The foundation is in place</span>
          </section>

          <section className="focus-cards" aria-label="Planning features">
            <FeatureCard
              icon={<Leaf size={18} />}
              title="Build steady habits"
              description="Small daily actions, made visible."
              status="Coming soon"
            />
            <FeatureCard
              icon={<Target size={18} />}
              title="Keep goals in view"
              description="Turn your quarter into a clear direction."
              status="Coming soon"
            />
            <FeatureCard
              icon={<Compass size={18} />}
              title="Find your weekly focus"
              description="A little intention for the days ahead."
              status="Coming soon"
            />
          </section>

          <footer className="planner-footer">
            <span>One step at a time.</span>
            <button
              type="button"
              className="quiet-signout"
              onClick={onLogout}
              disabled={isSigningOut}
            >
              Sign out <ArrowDownRight size={14} />
            </button>
          </footer>
        </div>
      </main>
    </div>
  );
}

function FeatureCard({
  icon,
  title,
  description,
  status,
}: {
  icon: ReactNode;
  title: string;
  description: string;
  status: string;
}) {
  return (
    <article className="feature-card">
      <span className="feature-icon">{icon}</span>
      <div>
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
      <span className="feature-status">{status}</span>
    </article>
  );
}
