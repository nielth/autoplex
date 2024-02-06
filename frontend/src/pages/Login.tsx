import { useEffect, useState } from "react";
import axios from "axios";
import { Box, Button, Container, Typography } from "@mui/material";
import { getWithExpiry, setWithExpiry } from "../scripts/localStorageExpire";

export default function LoginPage() {
  const [url, setUrl] = useState<string>("");
  const [error, setError] = useState<boolean>(false);

  useEffect(() => {
    (async () => {
      const localUrl = await getWithExpiry("authUrl");
      console.log(localUrl);
      if (localUrl === null) {
        axios.get("/api/authToken").then((resp) => {
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
              fontWeight: "900",
              fontSize: "10vw",
              position: "absolute",
              mt: 15,
              left: 0,
              right: 0,
              zIndex: -999,
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
              height={"100vh"}
              width={"400px"}
              alignItems={"center"}
              // bgcolor={"#0d1117"}
            >
              <Box>
                <Button
                  sx={{
                    color: "black",
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
