import { Link, useNavigate } from "react-router-dom";
import { DiskUsage } from "./DiskUsage";

export function Navbar() {
  const navigate = useNavigate();
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
          <Link to="/downloads" className="btn btn-ghost">
            Downloads
          </Link>
        </div>
      </div>
    </>
  );
}
