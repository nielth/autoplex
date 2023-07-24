import { Box, AppBar, Button, Avatar, Toolbar } from "@mui/material";
import { Outlet, Link } from "react-router-dom";
import { logout, oauthPlexLink, funcLoggedIn } from "../api/auth";
import { useEffect, useState } from "react";
import { getWithExpiry } from "../components/localStorExpire";

const arr = {
  Logs: "/logs",
  Downloads: "/downloads",
  // Home: "/",
};

async function setUrlFunc() {
  const url = await getWithExpiry("url");
  if (!url) {
    await oauthPlexLink();
  }
  return url;
}

export function NavBar() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [loading, setLoading] = useState(false);
  const [url, setUrl] = useState("");

  useEffect(() => {
    (async () => {
      await funcLoggedIn().then((resp) => {
        if (resp.data.logged_in_as) {
          setIsLoggedIn(true);
        } else {
          setUrlFunc().then((resp) => {
            setUrl(resp);
          });
        }
        setLoading(true);
      });
    })();
  }, [isLoggedIn, loading, url]);

  return (
    <>
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
            <Link to="/">
              <Button sx={{ minHeight: 1 }} size="large">
                <Avatar
                  variant="square"
                  alt="AP"
                  src="/AP_trans.png"
                  sx={{ mr: 3 }}
                />
                autoplex
              </Button>
            </Link>
            <Box sx={{ ml: "auto", mr: 0, height: "inherit" }}>
              {Object.entries(arr).map(([key, value]) => (
                <Link to={value}>
                  <Button sx={{ minHeight: 1 }} size="large">
                    {key}
                  </Button>
                </Link>
              ))}
              {
                isLoggedIn ?
                <Button
                onClick={async () => {
                  await logout();
                  setIsLoggedIn(false);
                }}
                >
                logout
              </Button> : <div />
              }
            </Box>
          </Toolbar>
        </AppBar>
      </Box>

      <Outlet context={{ isLoggedIn, setIsLoggedIn, loading, url }} />
    </>
  );
}
