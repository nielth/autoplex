import { Box, Button, CardMedia, Container, Typography } from "@mui/material";
import { oauthPlexLink } from "../lib/api";
import { getWithExpiry } from "../utils/localStorExpire";
import { useState } from "react";

async function setUrlFunc() {
  const url = await getWithExpiry("url");
  if (!url) {
    await oauthPlexLink();
  }
  return url;
}

// backgroundImage: 'url(https://media.giphy.com/media/3o7aDdlFRJU8D4F7QQ/giphy.gif);'
export function Login(context: any) {
  const [url, setUrl] = useState("");

  setUrlFunc().then((resp) => {
    setUrl(resp);
  });

  return (
    <>
      <Container
        maxWidth="lg"
        sx={{
          pt: { xs: "50px", lg: "100px" },
        }}
      >
        <Typography
          sx={{
            backgroundImage: "url('text-animation.gif')",
            textAlign: "center",
            backgroundSize: "cover",
            color: "transparent",
            backgroundClip: "text",
            fontWeight: "900",
            fontSize: { xs: "45px", lg: "100px" },
          }}
        >
          AUTOPLEX
        </Typography>

        <Box
          height={1}
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            width: {xs: '270px', lg: '320px'},
            backgroundColor: {xs: 'inherit', lg:"secondary.main"},
            border: "7px solid",
            borderWidth: "1px",
            borderColor: "border.main",
            minHeight: "300px",
            margin: "auto",
            borderRadius: "10px",
            mt: '70px'
          }}
        >
          <Button
            variant="outlined"
            onClick={function () {
              localStorage.removeItem("url");
            }}
            href={url}
            size="large"
            sx={{ color: "plex.main", borderColor: "border.main" }}
          >
            Continue with Plex
          </Button>
        </Box>
      </Container>
    </>
  );
}
