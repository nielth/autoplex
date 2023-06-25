import { Box, Button } from "@mui/material";
import axios from "axios";
import { useEffect, useState } from "react";
import { PAYLOAD } from "../components/payload";
import { getCookie } from "../components/getCookie";

export function Callback() {
  const [items, setItems] = useState([]);
  var url_auth = "https://plex.tv/api/v2/pins/";
  var url_auth_code = "";
  var authToken = "";

  useEffect(() => {
    const items = JSON.parse(localStorage.getItem("items") || "{}");
    PAYLOAD["X-Plex-Client-Identifier"] = items.clientIdentifier;
    url_auth_code = url_auth + items.id;
    console.log(url_auth_code);
    if (items) {
      setItems(items);
    }
  }, []);

  const plexInitAuth = () => {
    axios
      .get(url_auth_code, { headers: PAYLOAD })
      .then((res) => {
        authToken = res.data.authToken;
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
            axios.get("http://192.168.0.165:5000/protected", {
              withCredentials: true,
              headers: { "X-CSRF-TOKEN": getCookie("csrf_access_token") },
            });
          });
      })
      .catch((err) => {
        console.log(err);
      });
  };

  useEffect(() => {
    plexInitAuth();
  }, []);

  return <h1>asd</h1>;
}
