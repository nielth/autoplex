import axios from "axios";
import { PAYLOAD } from "../components/payload";
import { getCookie } from "../components/getCookies";
import { setWithExpiry, getWithExpiry } from "../components/localStorExpire";
import { v4 as uuidv4 } from "uuid";
import qs from "qs";

export async function login(url_auth_code: string) {
  await axios
    .get(url_auth_code, { headers: PAYLOAD })
    .then((res) => {
      const authToken = res.data.authToken;
      axios
        .post(
          "http://localhost:5000/authToken",
          { authToken: authToken },
          {
            withCredentials: true,
            headers: { "Access-Control-Allow-Origin": "*" },
          }
        )
        .then(() => {
          axios
            .get("http://localhost:5000/protected", {
              withCredentials: true,
              headers: { "X-CSRF-TOKEN": getCookie("csrf_access_token") },
            })
            .then(() => {
              return true;
            })
            .catch(() => {
              return false;
            });
        });
    })
    .catch((err) => {
      console.log(err);
    });
}

export async function logout() {
  await axios
    .post(
      "http://localhost:5000/logout",
      {},
      {
        withCredentials: true,
        headers: { "X-CSRF-TOKEN": getCookie("csrf_access_token") },
      }
    )
    .then(() => {

    })
    .catch(() => {
      console.log("error logout");
    });
}

export async function funcLoggedIn() {
  const resp = await axios
    .get("http://localhost:5000/protected", {
      withCredentials: true,
      headers: { "X-CSRF-TOKEN": getCookie("csrf_access_token") },
    })
    return resp
}

export async function oauthPlexLink() {
  const uuid4: string = uuidv4();
  var url_auth: string = "https://app.plex.tv/auth#!?";
  var forwardUrl: string = "http://localhost:3000/callback";
  var payload = PAYLOAD;
  payload["X-Plex-Client-Identifier"] = uuid4;
  var url_auth_para: string = "";

  await axios
    .post(
      "https://plex.tv/api/v2/pins.json?strong=true",
      qs.stringify(PAYLOAD),
      { headers: {} }
    )
    .then((res) => {
      const code = res.data.code;
      localStorage.setItem("items", JSON.stringify(res.data));
      url_auth_para =
        url_auth +
        "clientID=" +
        uuid4 +
        "&code=" +
        code +
        "&forwardUrl=" +
        forwardUrl;

      setWithExpiry("url", url_auth_para, res.data.expiresIn * 1000);
    })
    .catch(() => {
      return false;
    });

  return url_auth_para;
}
