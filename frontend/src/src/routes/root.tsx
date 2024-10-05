import { Outlet, useNavigate } from "react-router-dom";
import axios from "axios";

import { Navbar } from "../components/Navbar";
import { authProvider } from "../auth";
import { useEffect, useState } from "react";

export default function Root() {
  const [loading, setLoading] = useState<boolean>(true);
  const navigate = useNavigate();

  useEffect(() => {
    axios
      .get("http://localhost:5050/api/protected", {
        withCredentials: true,
      })
      .then((resp: any) => {
        console.log(resp.status);
        if (resp.status === 200 && resp.data.logged_in_as) {
          authProvider.signin(resp.data.logged_in_as);
          setLoading(false);
        }
      })
      .catch((err: any) => {
        if (err.response && err.response.status === 401) {
          authProvider.signout();
          navigate("/login");
        } else if (err.response) {
          navigate("/error");
        }
      });
  }, []);
  return (
    <>
          <Navbar />
      {!loading ? (
        <>
          <div id="detail" className="pt-8 container mx-auto">
            <Outlet />
          </div>
        </>
      ) : null}
    </>
  );
}
