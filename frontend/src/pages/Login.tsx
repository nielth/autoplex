import { useEffect, useState } from "react";
import axios from "axios";
import { Box, Button, Container, Typography } from "@mui/material";
import { getWithExpiry, setWithExpiry } from "../scripts/localStorageExpire";
import { useAuth } from "../App";
import { useNavigate } from "react-router-dom";

export default function LoginPage() {
  const [url, setUrl] = useState<string>("");
  const [error, setError] = useState<boolean>(false);
  const DOMAIN = process.env.REACT_APP_FLASK_LOCATION || "";
  let auth = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (auth.user) {
      navigate("/")
    }

    (async () => {
      const localUrl = await getWithExpiry("authUrl");
      if (localUrl === null) {
        axios
          .get(`${DOMAIN}/api/authToken`, { withCredentials: true })
          .then((resp) => {
            if (resp.status === 200) {
              setWithExpiry("authUrl", resp.data.url, 1800 * 1000);
              setUrl(resp.data.url);
            } else {
              setError(true);
            }
          });
      } else {
        setUrl(localUrl);
      }
    })();
  }, []);

  return (
    <>
      {url !== "" ? (
        <Box height={"100vh"} width={"100vw"}>
          <Typography
            sx={{
              backgroundImage: "url('text-animation.gif')",
              textAlign: "center",
              backgroundSize: "cover",
              color: "transparent",
              backgroundClip: "text",
              fontWeight: "bold",
              fontSize: { xs: "15vw", lg: "10vw" },
              position: "absolute",
              mt: 15,
              left: 0,
              right: 0,
              zIndex: -999,
              fontFamily: "Times New Roman",
            }}
          >
            AUTOPLEX
          </Typography>
          <Box display={"flex"} justifyContent={"center"}>
            <Box
              display={"flex"}
              justifyContent={"center"}
              // border={"2px solid #30363d"}
              borderRadius={"6px"}
              mt={"500px"}
              width={"400px"}
              alignItems={"center"}
            // bgcolor={"#0d1117"}
            >
              <Box>
                <Button
                  sx={{
                    color: "black",
                    backgroundColor: "#e5a00d",
                    "&:hover": {
                      backgroundColor: "inherit",
                      color: "#e5a00d",
                      border: "1px solid #30363d",
                    },
                  }}
                  onClick={() => {
                    localStorage.removeItem("authUrl");
                    window.location.href = url;
                  }}
                >
                  Continue with Plex
                </Button>
              </Box>
            </Box>
          </Box>
        </Box>
      ) : null}
      {error ? (
        <>
          <Container>
            <Box mt={25}>
              <Typography textAlign={"center"} color={"white"}>
                Something went wrong. Try again in 5 minutes.
                <br />
                If prolonged, contact admin.
              </Typography>
            </Box>
          </Container>
        </>
      ) : null}
    </>
  );
}
