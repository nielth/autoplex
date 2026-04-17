import { useEffect, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { DiskUsage } from "./DiskUsage";

function formatUtcNow(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, "0");
  return (
    `${date.getUTCFullYear()}-${pad(date.getUTCMonth() + 1)}-${pad(date.getUTCDate())} ` +
    `${pad(date.getUTCHours())}:${pad(date.getUTCMinutes())}:${pad(date.getUTCSeconds())} UTC`
  );
}

export function Navbar() {
  const navigate = useNavigate();
  const location = useLocation();
  const [mobileMenuOpen, setMobileMenuOpen] = useState<boolean>(false);
  const [utcNow, setUtcNow] = useState<string>(() => formatUtcNow(new Date()));

  const closeMobileMenu = () => {
    setMobileMenuOpen(false);
  };

  useEffect(() => {
    setMobileMenuOpen(false);
  }, [location.pathname, location.search, location.hash]);

  useEffect(() => {
    const id = window.setInterval(() => {
      setUtcNow(formatUtcNow(new Date()));
    }, 1000);
    return () => window.clearInterval(id);
  }, []);

  return (
    <>
      <div className="navbar bg-base-300">
        <div className="navbar-start">
          <div className="flex items-center">
            <Link to="/">
              <button
                className="btn btn-ghost text-xl"
                onClick={() => {
                  navigate("/");
                }}
              >
                Autoplex
              </button>
            </Link>
          </div>
        </div>
        <DiskUsage />
        <div className="navbar-end">
          <div className="dropdown dropdown-end lg:hidden">
            <button
              type="button"
              role="button"
              aria-expanded={mobileMenuOpen}
              className="btn btn-ghost btn-circle"
              onClick={() => {
                setMobileMenuOpen((previous) => !previous);
              }}
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                className="inline-block h-5 w-5 stroke-current"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M4 6h16M4 12h16M4 18h16"
                />
              </svg>
            </button>
          </div>
          {mobileMenuOpen ? (
            <div className="fixed inset-0 z-[100] bg-base-100 lg:hidden">
              <button
                type="button"
                className="btn btn-ghost btn-circle absolute right-4 top-4"
                onClick={closeMobileMenu}
                aria-label="Close menu"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  className="inline-block h-5 w-5 stroke-current"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M6 6l12 12M18 6l-12 12"
                  />
                </svg>
              </button>
              <ul className="menu menu-lg m-0 flex h-full w-full flex-col items-center justify-center gap-6 p-0 text-center">
                <li className="w-full max-w-sm">
                  <Link
                    to="/"
                    className="justify-center text-3xl font-semibold"
                    onClick={closeMobileMenu}
                  >
                    Home
                  </Link>
                </li>
                <li className="w-full max-w-sm">
                  <Link
                    to="/tvmaze"
                    className="justify-center text-3xl font-semibold"
                    onClick={closeMobileMenu}
                  >
                    TV Scheduler
                  </Link>
                </li>
                <li className="w-full max-w-sm">
                  <Link
                    to="/downloads"
                    className="justify-center text-3xl font-semibold"
                    onClick={closeMobileMenu}
                  >
                    Downloads
                  </Link>
                </li>
              </ul>
            </div>
          ) : null}
          <div className="hidden lg:flex items-center">
            <span
              className="mr-3 font-mono text-sm opacity-70 tabular-nums"
              title="Current UTC time"
            >
              {utcNow}
            </span>
            <ul className="menu menu-horizontal px-1">
              <li>
                <Link to="/">Home</Link>
              </li>
              <li>
                <Link to="/tvmaze">TV Scheduler</Link>
              </li>
              <li>
                <Link to="/downloads">Downloads</Link>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </>
  );
}
