import { Box, Button } from "@mui/material";
import { v4 as uuidv4 } from "uuid";
import axios from "axios";
import qs from "qs";
import { useEffect, useState } from "react";
import { PAYLOAD } from "../components/payload";

export function PageA() {
  const uuid4 = uuidv4();
  const identifier = "X-Plex-Client-Identifier";
  PAYLOAD[identifier] = uuid4;
  const [url, setUrl] = useState("");
  var url_auth = "https://app.plex.tv/auth#!?";
  var forwardUrl = "http://localhost:3000/callback";
  var code = "";
  const plexInitAuth = () => {
    axios
      .post(
        "https://plex.tv/api/v2/pins.json?strong=true",
        // "http://192.168.0.165:5000/api",
        qs.stringify(PAYLOAD),
        { headers: {} }
      )
      .then((res) => {
        console.log(res.data);
        code = res.data.code;
        localStorage.setItem("items", JSON.stringify(res.data));
        var url_auth_para =
          url_auth +
          "clientID=" +
          uuid4 +
          "&code=" +
          code +
          "&forwardUrl=" +
          forwardUrl;
        setUrl(url_auth_para);
        console.log(url_auth_para);
      })
      .catch((err) => {
        console.log(err);
      });
  };

  useEffect(() => {
    plexInitAuth();
  }, []);

  return (
    <div>
      <div className="PageContainer">
        <Box
          height={1}
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            width: "300px",
            backgroundColor: "#171b21",
            borderColor: "#202429",
            borderRadius: "7px",
            borderStyle: "solid",
            borderWidth: "1px",
            minWidth: "400px",
            minHeight: "400px",
            margin: "auto",
          }}
        >
          <Button variant="outlined" href={url} size="large" color="plex_col">
            Continue with Plex
          </Button>
        </Box>
      </div>
    </div>
  );
}
