import { Box, Button } from "@mui/material";
import axios from "axios";
import { useEffect, useState } from "react";
import { PAYLOAD } from "../components/payload";

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
        localStorage.setItem("authToken", res.data.authToken);
        authToken = res.data.authToken;
        axios.post(
          "http://192.168.0.165:5000/authToken",
          { authToken: authToken },
          {
            headers: {},
          }
        );
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
