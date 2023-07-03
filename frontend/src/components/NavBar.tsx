import { Box, AppBar, Button, Avatar, Toolbar } from "@mui/material";
import { Outlet, Link } from "react-router-dom";
import { logout } from "../api/auth";
import { useNavigate } from "react-router-dom";

const arr = {
  Logs: "/logs",
  Downloads: "/downloads",
  Home: "/",
};

export function NavBar() {
  const navigate = useNavigate();

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
              <Button onClick={
                async () => {
                  await logout()
                  window.location.reload();
                }
              }>logout</Button>
            </Box>
          </Toolbar>
        </AppBar>
      </Box>

      <Outlet />
    </>
  );
}
