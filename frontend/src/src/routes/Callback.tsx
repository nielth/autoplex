import { useEffect, useState } from "react";
import { Navigate } from "react-router-dom";
import axios from "axios";

import { authProvider } from "../auth";

export default function Callback() {
  const [fin, setFin] = useState<boolean>(false);

  useEffect(() => {
    axios
      .get("http://localhost:5050/api/callback", {
        withCredentials: true,
      })
      .then((resp: any) => {
        if (resp.status === 200 && resp.data.logged_in_as) {
          authProvider.signin(resp.data.logged_in_as);
        }
      })
      .finally(() => {
        setFin(true);
      });
  }, []);
  return (
    <>
      <div className="container pt-16 mx-auto text-center">
        <span className="loading loading-spinner loading-lg"></span>
        {fin ? (
          <>
            <Navigate to="/" />
          </>
        ) : null}
      </div>
    </>
  );
}
