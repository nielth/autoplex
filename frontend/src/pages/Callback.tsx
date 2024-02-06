import { Box, CircularProgress, Container, Typography } from "@mui/material";
import axios from "axios";
import { useEffect, useState } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { useAuth } from "../App";

export default function Callback() {
  const [error, setError] = useState<number>(0);
  const navigate = useNavigate();
  let auth = useAuth();

  useEffect(() => {
    axios
      .get("/api/callback")
      .then(
        (resp: {
          status: number;
          data: { logged_in_as: string; logged_in: boolean };
        }) => {
          if (resp.data.logged_in === true) {
            auth.setUser(resp.data.logged_in_as);
          }
          navigate("/");
        }
      )
      .catch((error: { status: number }) => {
        setError(error.status);
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
        {error === 401 ? (
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
