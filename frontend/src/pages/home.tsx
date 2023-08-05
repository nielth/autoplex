import { useEffect, useState } from "react";
import {
  AppBar,
  Avatar,
  Box,
  Button,
  Container,
  IconButton,
  Menu,
  MenuItem,
  Toolbar,
  Tooltip,
  Typography,
} from "@mui/material";
import MenuIcon from "@mui/icons-material/Menu";
import { funcLoggedIn } from "../lib/api";
import { Navigate, redirect } from "react-router-dom";

declare module "@mui/material/AppBar" {
  interface AppBarPropsColorOverrides {
    navbar: true;
  }
}

const pages = ["Home", "Downloads", "Logs"];
const settings = ["Profile", "Logout"];

export function ResponsiveAppBar() {
  const [anchorElNav, setAnchorElNav] = useState<null | HTMLElement>(null);
  const [anchorElUser, setAnchorElUser] = useState<null | HTMLElement>(null);

  const handleOpenNavMenu = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorElNav(event.currentTarget);
  };
  const handleOpenUserMenu = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorElUser(event.currentTarget);
  };

  const handleCloseNavMenu = () => {
    setAnchorElNav(null);
  };

  const handleCloseUserMenu = () => {
    setAnchorElUser(null);
  };

  return (
    <>
      <AppBar
        color="navbar"
        sx={{ borderBottom: "1px solid", borderColor: "border.main" }}
        position="static"
      >
        <Container maxWidth="xl">
          <Toolbar disableGutters>
            <Button
              href="/"
              sx={{ display: { xs: "none", md: "flex" }, minHeight: 1 }}
              size="large"
            >
              <Avatar
                variant="square"
                alt="AP"
                src="/AP_trans.png"
                sx={{ mr: 3 }}
              />
              <Typography
                variant="h6"
                noWrap
                component="a"
                sx={{
                  mr: 1,
                  display: { xs: "none", md: "flex" },

                  fontWeight: 700,
                  textDecoration: "none",
                }}
              >
                Autoplex
              </Typography>
            </Button>

            <Box sx={{ flexGrow: 1, display: { xs: "flex", md: "none" } }}>
              <IconButton
                size="large"
                aria-label="account of current user"
                aria-controls="menu-appbar"
                aria-haspopup="true"
                onClick={handleOpenNavMenu}
                color="inherit"
              >
                <MenuIcon />
              </IconButton>
              <Menu
                id="menu-appbar"
                anchorEl={anchorElNav}
                anchorOrigin={{
                  vertical: "bottom",
                  horizontal: "left",
                }}
                keepMounted
                transformOrigin={{
                  vertical: "top",
                  horizontal: "left",
                }}
                open={Boolean(anchorElNav)}
                onClose={handleCloseNavMenu}
                sx={{
                  display: { xs: "block", md: "none" },
                }}
              >
                {pages.map((page) => (
                  <MenuItem
                    key={page}
                    sx={{ color: "black" }}
                    onClick={handleCloseNavMenu}
                  >
                    <Typography textAlign="center">{page}</Typography>
                  </MenuItem>
                ))}
              </Menu>
            </Box>
            <Container
              sx={{ flexGrow: 2, display: { xs: "flex", md: "none" } }}
            >
              <Button sx={{ m: "auto", textAlign: "center" }}>
                <Avatar
                  variant="square"
                  alt="AP"
                  src="/AP_trans.png"
                  sx={{ mr: 2 }}
                />
                <Typography
                  variant="h6"
                  noWrap
                  component="a"
                  sx={{
                    mr: 1,
                    display: { xs: "flex", md: "none" },
                    flexGrow: 1,
                    fontWeight: 700,
                    textDecoration: "none",
                  }}
                >
                  Autoplex
                </Typography>
              </Button>
            </Container>
            <Box sx={{ flexGrow: 1, display: { xs: "none", md: "flex" } }}>
              {pages.map((page) => (
                <Button
                  key={page}
                  onClick={handleCloseNavMenu}
                  sx={{ my: 2, color: "white", display: "block" }}
                >
                  {page}
                </Button>
              ))}
            </Box>
            <Box sx={{ flexGrow: 0 }}>
              <Tooltip title="Open settings">
                <IconButton onClick={handleOpenUserMenu} sx={{ p: 0 }}>
                  <Avatar alt="Thomas" src="/static/images/avatar/2.jpg" />
                </IconButton>
              </Tooltip>
              <Menu
                sx={{ mt: "45px" }}
                id="menu-appbar"
                anchorEl={anchorElUser}
                anchorOrigin={{
                  vertical: "top",
                  horizontal: "right",
                }}
                keepMounted
                transformOrigin={{
                  vertical: "top",
                  horizontal: "right",
                }}
                open={Boolean(anchorElUser)}
                onClose={handleCloseUserMenu}
              >
                {settings.map((setting) => (
                  <MenuItem
                    key={setting}
                    onClick={handleCloseUserMenu}
                    sx={{ color: "black" }}
                  >
                    <Typography textAlign="center">{setting}</Typography>
                  </MenuItem>
                ))}
              </Menu>
            </Box>
          </Toolbar>
        </Container>
      </AppBar>
      <Container maxWidth="xl">
        <Typography>Hello</Typography>
      </Container>
    </>
  );
}

export function Home() {
  const [loading, setLoading] = useState(true);
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    setLoading(true)
    funcLoggedIn().then((resp) => {
      if (resp.data.logged_in_as) {
        console.log(resp.data.logged_in_as)
        setIsLoggedIn(true);
      } else {
      }
      setLoading(false);
    });
  }, []);

  return (
    <>
      {loading ? (
        <div />
      ) : isLoggedIn ? (
        <ResponsiveAppBar />
      ) : (
        <Navigate to={"/login"} replace={true} />
      )}
    </>
  );
}
