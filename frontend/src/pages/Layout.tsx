import { Outlet, useNavigate } from "react-router-dom";
import AppBar from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import Toolbar from "@mui/material/Toolbar";
import Container from "@mui/material/Container";
import Button from "@mui/material/Button";
import axios from "axios";
import { useAuth } from "../App";

function ResponsiveAppBar() {
  const navigate = useNavigate();
  let auth = useAuth();

  return (
    <AppBar position="static">
      <Container maxWidth="xl">
        <Toolbar disableGutters>
          <Box position={"absolute"} width={"100%"} textAlign={"center"}>
            <Button
              disableElevation
              disableFocusRipple
              disableRipple
              disableTouchRipple
              onClick={() => {
                navigate("/");
              }}
            >
              <Box component={"img"} src="/AP_trans.png" width={"61px"} />
            </Button>
          </Box>
          <Box
            justifyContent={"flex-end"}
            sx={{ flexGrow: 1, display: "flex" }}
          >
            <Button
              key={"logout"}
              sx={{ my: 2, color: "white", display: "block" }}
              onClick={() => {
                auth.signout(() => {
                  navigate("/");
                });
              }}
            >
              Logout
            </Button>
          </Box>
        </Toolbar>
      </Container>
    </AppBar>
  );
}

export default function Layout() {
  return (
    <>
      <ResponsiveAppBar />
      <Container>
        <Outlet />
      </Container>
    </>
  );
}
