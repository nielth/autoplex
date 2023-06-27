import { Box, Button } from "@mui/material";
import { v4 as uuidv4 } from "uuid";
import axios from "axios";
import qs from "qs";
import { useEffect, useState } from "react";
import { PAYLOAD } from "../components/payload";
import { setWithExpiry, getWithExpiry } from "../components/localStorExpire";

function notLoggedIn(url: string) {
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

export function PageA() {
  const [url, setUrl] = useState("");
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [loading, setLoading] = useState(false);
  const [posts, setPosts] = useState([]);

  useEffect(() => {
    async function verifyAuth() {
      const resp = await axios
        .get("http://192.168.0.165:5000/protected", {
          withCredentials: true,
        })
        .then(() => {
          setIsLoggedIn(true);
        })
        .catch(() => {
          setIsLoggedIn(false);
        });
        console.log("1")
    }

    const loadPost = async () => {
      // Till the data is fetch using API
      // the Loading page will show.
      setLoading(true);

      const uuid4 = uuidv4();
      PAYLOAD["X-Plex-Client-Identifier"] = uuid4;
      var url_auth = "https://app.plex.tv/auth#!?";
      var forwardUrl = "http://localhost:3000/callback";

      // Await make wait until that
      // promise settles and return its result
      const resp: any = await axios
        .post(
          "https://plex.tv/api/v2/pins.json?strong=true",
          qs.stringify(PAYLOAD),
          { headers: {} }
        )
        .then((res) => {
          var code = res.data.code;
          localStorage.setItem("items", JSON.stringify(res.data));
          setPosts(res.data);
          var url_auth_para =
            url_auth +
            "clientID=" +
            uuid4 +
            "&code=" +
            code +
            "&forwardUrl=" +
            forwardUrl;
          setUrl(url_auth_para);
          setWithExpiry("url", url_auth_para, res.data.expiresIn * 1000);
        })
        .catch(() => {});

      setLoading(false);
    };

    const buttonUrl = getWithExpiry("url");

    if (!buttonUrl) {
      loadPost();
    } else {
      setUrl(buttonUrl);
    }

    verifyAuth().then(() => {
      setLoading(true);
    })
    
  }, []);

  useEffect(() => {
    const interval = setInterval(() => {
      console.log("test");
    }, 2000);

    return () => clearInterval(interval);
  }, []);

  return (
    <>{loading ? (isLoggedIn ? <h1>Logged in</h1> : url ? notLoggedIn(url) : <div />) : <div />}</>
  );
}
