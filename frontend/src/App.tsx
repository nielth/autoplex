import AppBar from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import Toolbar from "@mui/material/Toolbar";
import Button from "@mui/material/Button";
import Avatar from "@mui/material/Avatar";
import { createTheme, ThemeProvider } from "@mui/material/styles";
import {
  BrowserRouter as Router,
  Route,
  Routes,
  BrowserRouter,
} from "react-router-dom";
import { PageA } from "./pages/home";
import { PageB } from "./pages/pageB";
import { Callback } from "./pages/callback";

var arr = {
  Logs: "/logs",
  Downloads: "/downloads",
  Home: "/"
};

const theme = createTheme({
  palette: {
    primary: {
      main: "#e7edf2",
    },
    plex_col: {
      main: "#c17f34",
    },
  },
  typography: {
    // In Chinese and Japanese the characters are usually larger,
    // so a smaller fontsize may be appropriate.
    button: {
      fontWeight: "bold",
    },
    fontSize: 17,
    fontFamily: ["Tahoma", "Verdana", "Arial", "serif"].join(","),
  },
});

declare module "@mui/material/styles" {
  interface Palette {
    plex_col: Palette["primary"];
  }

  // allow configuration using `createTheme`
  interface PaletteOptions {
    plex_col?: PaletteOptions["primary"];
  }
}

// @babel-ignore-comment-in-output Update the Button's color prop options
declare module "@mui/material/Button" {
  interface ButtonPropsColorOverrides {
    plex_col: true;
  }
}

function App() {
  return (
    <>
      <ThemeProvider theme={theme}>
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
              <Button sx={{ minHeight: 1 }} size="large" href="/">
                <Avatar
                  variant="square"
                  alt="AP"
                  src="/AP_trans.png"
                  sx={{ mr: 3 }}
                />
                autoplex
              </Button>
              <Box sx={{ ml: "auto", mr: 0, height: "inherit" }}>
                {Object.entries(arr).map(([key, value]) => (
                  <Button sx={{ minHeight: 1 }} size="large" href={value}>
                    {key}
                  </Button>
                ))}
              </Box>
            </Toolbar>
          </AppBar>
        </Box>

        <BrowserRouter>
          <Routes>
            <Route path="/" element={<PageA />} />
            <Route path="/downloads" element={<PageB />} />
            <Route path="/callback" element={<Callback />} />
          </Routes>
        </BrowserRouter>
      </ThemeProvider>
    </>
  );
}

export default App;
