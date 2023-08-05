import { useEffect } from "react";
import { login } from "../lib/api";
import { useNavigate } from "react-router-dom";

export function Callback() {
  const navigate = useNavigate();

  useEffect(() => {
    (async () => {
      let url_auth = "https://plex.tv/api/v2/pins/";
      let url_auth_code = "";
      const lstore = JSON.parse(localStorage.getItem("response") || "{}");
      const payload = JSON.parse(localStorage.getItem("payload") || "{}");
      payload["X-Plex-Client-Identifier"] = lstore.clientIdentifier;
      url_auth_code = url_auth + lstore.id;

      await login(url_auth_code, payload).then((answer) => {
        console.log(answer);
        navigate("/", { replace: true });
      });
    })();
  }, []);

  return <div>Redirecting...</div>;
}
