import "./styles.css";
import AppBar from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import Toolbar from "@mui/material/Toolbar";
import Button from "@mui/material/Button";
import Avatar from "@mui/material/Avatar";
import { createTheme, ThemeProvider } from "@mui/material/styles";

import { v4 as uuidv4 } from "uuid";
import axios from "axios";
import qs from "qs";
import { useEffect, useState } from "react";

var arr = {
  Logs: "/logs",
  Downloads: "/downloads",
};

const theme = createTheme({
  palette: {
    primary: {
      main: "#e7edf2",
    },
    plex_col: {
      main: "#c17f34",
    },
  },
  typography: {
    // In Chinese and Japanese the characters are usually larger,
    // so a smaller fontsize may be appropriate.
    button: {
      fontWeight: "bold",
    },
    fontSize: 17,
    fontFamily: ["Tahoma", "Verdana", "Arial", "serif"].join(","),
  },
});

declare module "@mui/material/styles" {
  interface Palette {
    plex_col: Palette["primary"];
  }

  // allow configuration using `createTheme`
  interface PaletteOptions {
    plex_col?: PaletteOptions["primary"];
  }
}

// @babel-ignore-comment-in-output Update the Button's color prop options
declare module "@mui/material/Button" {
  interface ButtonPropsColorOverrides {
    plex_col: true;
  }
}

function App() {
  const uuid4 = uuidv4();

  var PAYLOAD = {
    "X-Plex-Product": "Plex Auth App (Autoplex)",
    "X-Plex-Version": "0.69.420",
    "X-Plex-Device": "Linux",
    "X-Plex-Platform": "Linux",
    "X-Plex-Device-Name": "Autoplex",
    "X-Plex-Device-Vendor": "",
    "X-Plex-Model": "",
    "X-Plex-Client-Platform": "",
    "X-Plex-Client-Identifier": uuid4,
  };

  const [url, setUrl] = useState("");
  var url_auth = "https://app.plex.tv/auth#!?";
  var forwardUrl = "http://localhost:5000";
  var code = "";
  var identifier = "";
  const plexInitAuth = () => {
    axios
      .post(
        // "https://plex.tv/api/v2/pins.json?strong=true",
        "http://192.168.0.165:5000/api",
        qs.stringify(PAYLOAD),
        { headers: {} }
      )
      .then((res) => {
        console.log(res.data);
        code = res.data.code;
        identifier = res.data.id;
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
    <>
      <ThemeProvider theme={theme}>
        <Box
          sx={{
            bgcolor: "#020409",
            borderStyle: "solid",
            border: 0,
            borderBottom: 1,
            borderColor: "#202429",
          }}
        >
          <AppBar position="static" elevation={0} sx={{ bgcolor: "inherit" }}>
            <Toolbar sx={{ height: 70 }}>
              <Button sx={{ minHeight: 1 }} size="large" href="/">
                <Avatar
                  variant="square"
                  alt="AP"
                  src="/AP_trans.png"
                  sx={{ mr: 3 }}
                />
                autoplex
              </Button>
              <Box sx={{ ml: "auto", mr: 0, height: "inherit" }}>
                {Object.entries(arr).map(([key, value]) => (
                  <Button sx={{ minHeight: 1 }} size="large" href={value}>
                    {key}
                  </Button>
                ))}
              </Box>
            </Toolbar>
          </AppBar>
        </Box>
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
      </ThemeProvider>
    </>
  );
}

export default App;
