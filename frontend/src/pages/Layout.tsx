import { Outlet, useNavigate } from "react-router-dom";
import AppBar from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import Toolbar from "@mui/material/Toolbar";
import Container from "@mui/material/Container";
import Button from "@mui/material/Button";
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
              href="/"
            >
              <Box component={"img"} src="/AP_trans.png" width={"61px"} />
            </Button>
          </Box>
          <Box display={"flex"}>
            <Box>
              <Button
                key={"home"}
                sx={{ my: 2, color: "white", display: "block" }}
                href="/"
              >
                Home
              </Button>
            </Box>
            <Box>
              <Button
                key={"rss"}
                sx={{ my: 2, color: "white", display: "block" }}
                onClick={() => {
                  navigate("/rss");
                }}
                disabled
              >
                RSS
              </Button>
            </Box>
          </Box>
          <Box
            justifyContent={"flex-end"}
            sx={{ display: "flex" }}
            width={"100%"}
          >
            <Box>
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
      <Container maxWidth="xl" sx={{ py: 5, textAlign: "center" }}>
        <Outlet />
      </Container>
    </>
  );
}
