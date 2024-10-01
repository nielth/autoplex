import { useEffect, useState } from "react";
import { getWithExpiry, setWithExpiry } from "../scripts/localStorage";
import axios from "axios";

export function Login() {
  const [url, setUrl] = useState<string>("");

  useEffect(() => {
    (async () => {
      const localUrl = await getWithExpiry("authUrl");
      if (localUrl === null) {
        axios
          .get(`https://autoplex.nielth.com/api/authToken`, {
            withCredentials: true,
          })
          .then((resp) => {
            if (resp.status === 200) {
              setWithExpiry("authUrl", resp.data.url, 1800 * 1000);
              setUrl(resp.data.url);
            } else {
            }
          });
      } else {
        setUrl(localUrl);
      }
    })();
  }, []);

  return (
    <>
      <div className="pt-8">
        <div className="text-center">
          {url !== "" ? (
            <>
              <p className="text-7xl">Welcome to Autoplex</p>
              <button
                className="btn mt-16"
                onClick={() => {
                  localStorage.removeItem("authUrl");
                  window.location.href = url;
                }}
              >
                Continue with Plex
              </button>
            </>
          ) : (
            <>
              <span className="loading loading-spinner loading-lg"></span>
            </>
          )}
        </div>
      </div>
    </>
  );
}
