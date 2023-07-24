import { useEffect, useState } from "react";
import { login } from "../api/auth";
import { PAYLOAD } from "../components/payload";
import { useNavigate, useOutletContext } from "react-router-dom";
import { Dict } from "styled-components/dist/types";

export function Callback() {
  var url_auth = "https://plex.tv/api/v2/pins/";
  var url_auth_code = "";

  const navigate = useNavigate();
  const context: Dict = useOutletContext();

  useEffect(() => {
    (async () => {
      const lstore = JSON.parse(localStorage.getItem("items") || "{}");
      PAYLOAD["X-Plex-Client-Identifier"] = lstore.clientIdentifier;
      url_auth_code = url_auth + lstore.id;
      await login(url_auth_code).then(() => {
        context.setIsLoggedIn(true);
        navigate("/");
      });
    })();
  }, []);

  return <div>Redirecting...</div>;
}
