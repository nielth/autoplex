import { useEffect, useState } from "react";
import { login } from "../api/auth";
import { PAYLOAD } from "../components/payload";
import { useNavigate } from "react-router-dom";

export function Callback() {
  const [success, setSuccess] = useState(false);
  const [loaded, setLoaded] = useState(false);
  var url_auth = "https://plex.tv/api/v2/pins/";
  var url_auth_code = "";

  const navigate = useNavigate();

  useEffect(() => {
    (async () => {
      const lstore = JSON.parse(localStorage.getItem("items") || "{}");
      PAYLOAD["X-Plex-Client-Identifier"] = lstore.clientIdentifier;
      url_auth_code = url_auth + lstore.id;
      console.log(url_auth_code);
      login(url_auth_code).then(() => {
        console.log("logged in")
        setSuccess(true);
        setLoaded(true)
      });
    })();
  }, []);

  return <>{loaded ? (success ? navigate("/") : <h1>not</h1>) : <div />}</>;
}
