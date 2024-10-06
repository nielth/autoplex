import { Outlet, useNavigate } from "react-router-dom";
import axios from "axios";

import { Navbar } from "../components/Navbar";
import { authProvider } from "../auth";
import { useEffect, useState } from "react";
import { getApiDomain } from "../scripts/getApiDomain";

export default function Root() {
  const [loading, setLoading] = useState<boolean>(true);
  const navigate = useNavigate();
  const domain = getApiDomain();

  useEffect(() => {
    axios
      .get(`${domain}/api/protected`, {
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
      {!loading ? (
        <>
          <Navbar />
          <div id="detail" className="py-8 lg:container mx-auto">
            <Outlet />
          </div>
        </>
      ) : null}
    </>
  );
}
