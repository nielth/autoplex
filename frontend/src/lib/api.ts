import axios from "axios";
import { v4 as uuidv4 } from "uuid";
import qs from "qs";

import { getCookie } from "../utils/getCookies";
import { setWithExpiry } from "../utils/localStorExpire";

export async function login(url_auth_code: string, payload: any) {
  let res = await axios.get(url_auth_code, { headers: payload });

  if (res.status !== 200) {
    return null;
  }

  const authToken = res.data.authToken;
  console.log(authToken);
  let res2 = await axios.post(
    "http://localhost:5000/authToken",
    { authToken: authToken },
    {
      withCredentials: true,
      headers: { "Access-Control-Allow-Origin": "*" },
    }
  );

  if (res2.status !== 200) {
    return null;
  }

  return true;
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
    .then(() => {})
    .catch(() => {
      console.log("error logout");
    });
}

export async function funcLoggedIn() {
  const resp = await axios.get("http://localhost:5000/protected", {
    withCredentials: true,
    headers: { "X-CSRF-TOKEN": getCookie("csrf_access_token") },
  });
  return resp;
}

export async function oauthPlexLink() {
  let payload = {
    "X-Plex-Product": "Plex Auth App (Autoplex)",
    "X-Plex-Version": "0.69.420",
    "X-Plex-Device": "Linux",
    "X-Plex-Platform": "Linux",
    "X-Plex-Device-Name": "Autoplex",
    "X-Plex-Device-Vendor": "",
    "X-Plex-Model": "",
    "X-Plex-Client-Platform": "",
    "X-Plex-Client-Identifier": "",
    "Content-Type": "application/json",
    Accept: "application/json",
  };

  const uuid4: string = uuidv4();
  payload["X-Plex-Client-Identifier"] = uuid4;
  let url_auth: string = "https://app.plex.tv/auth#!?";
  let forwardUrl: string = "http://localhost:3000/callback";
  let url_auth_para: string = "";

  await axios
    .post(
      "https://plex.tv/api/v2/pins.json?strong=true",
      qs.stringify(payload),
      { headers: {} }
    )
    .then((res) => {
      const code = res.data.code;
      localStorage.setItem("response", JSON.stringify(res.data));
      localStorage.setItem("payload", JSON.stringify(payload));
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

export async function torrentPost(data: any) {
  await axios
    .post(
      "http://localhost:5000/download",
      { fid: data.fid, filename: data.filename, category: data.categoryID },
      {
        withCredentials: true,
        headers: { "X-CSRF-TOKEN": getCookie("csrf_access_token") },
      }
    )
    .then((res) => {
      return true;
    })
    .catch(() => {
      return false;
    });
}

export async function torrentSearch(data: any) {
  const encodeData = encodeURI(data);
  const response = await axios.post(
    "http://localhost:5000/search",
    { search: encodeData },
    {
      withCredentials: true,
      headers: { "X-CSRF-TOKEN": getCookie("csrf_access_token") },
    }
  );

  return response.data;
}
