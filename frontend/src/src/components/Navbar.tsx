import { Link, useNavigate } from "react-router-dom";

export function Navbar() {
  const navigate = useNavigate();
  return (
    <>
      <div className="navbar bg-base-300">
        <div className="navbar-start">
          <Link to="/">
            <button
              className="btn btn btn-ghost text-xl"
              onClick={() => {
                navigate("/");
              }}
            >
              Autoplex
            </button>
          </Link>
        </div>
        <div className="navbar-end">
          <Link to="/login">
            <button className="btn btn-ghost">Login</button>
          </Link>
        </div>
      </div>
    </>
  );
}
