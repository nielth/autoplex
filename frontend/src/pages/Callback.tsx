import { Box, CircularProgress, Container, Typography } from "@mui/material";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../App";

export default function Callback() {
  const [error, setError] = useState<number>(0);
  const navigate = useNavigate();
  let auth = useAuth();

  useEffect(() => {
    auth.signin((resp) => {
      if (resp === 0) {
        navigate("/");
      } else {
        setError(resp);
      }
    });
  }, []);
  return (
    <>
      {error === 0 ? (
        <Box sx={{ display: "flex" }}>
          <CircularProgress />
        </Box>
      ) : null}

      <Box component={Container} mt={25} color={"white"}>
        {error === 403 ? (
          <>
            {((): any => {
              setTimeout(() => {
                navigate("/");
              }, 5000);
            })()}
            <Typography textAlign={"center"}>
              AUTHORIZATION FAILED. Newly added to server? Wait 24 hours.
            </Typography>
          </>
        ) : error === 500 ? (
          <>
            {((): any => {
              setTimeout(() => {
                navigate("/");
              }, 5000);
            })()}
            <Typography textAlign={"center"}>
              SERVER ERROR. Contact admin if persistent.
            </Typography>
          </>
        ) : null}
      </Box>
    </>
  );
}
