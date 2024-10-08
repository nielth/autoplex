import { Link, redirect, useNavigate } from "react-router-dom";
import { authProvider } from "../auth";
import axios from "axios";
import { getApiDomain } from "../scripts/getApiDomain";

async function logout() {
  const domain = getApiDomain();
  try {
    const resp = await axios.get(`${domain}/api/logout`, {
      withCredentials: true,
    });
    if (resp && resp.status === 200) {
      return true;
    }
  } catch (error) {
    return false;
  }
}

export function Navbar() {
  const navigate = useNavigate();
  return (
    <>
      <div className="navbar bg-base-300">
        <div className="navbar-start">
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
        <div className="navbar-end">
          <button
            className="btn btn-ghost"
            onClick={() => {
              logout().then((resp) => {
                if (resp) {
                  authProvider.signout();
                  navigate("/login");
                } else {
                  navigate("/error");
                }
              });
            }}
          >
            Logout
          </button>
        </div>
      </div>
    </>
  );
}
