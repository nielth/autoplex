import axios from "axios";
import { PAYLOAD } from "../components/payload";

export async function login(url_auth_code: string) {
  await axios
    .get(url_auth_code, { headers: PAYLOAD })
    .then((res) => {
      const authToken = res.data.authToken;
      axios
        .post(
          "http://192.168.0.165:5000/authToken",
          { authToken: authToken },
          {
            withCredentials: true,
            headers: { "Access-Control-Allow-Origin": "*" },
          }
        )
        .then(() => {
          axios
            .get("http://192.168.0.165:5000/protected", {
              withCredentials: true,
            })
            .then(() => {
              console.log("true, king");
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
  (async () => {
    const resp = await axios.post("http://192.168.0.165:5000/logout", {
      withCredentials: true,
    });
  })();
}
