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
import { logout } from "./api/auth";
import { NavBar } from "./components/NavBar";

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

export async function temp() {}

export default function App() {
  return (
    <>
      <ThemeProvider theme={theme}>
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<NavBar />}>
              <Route index element={<PageA />} />
              <Route path="downloads" element={<PageB />} />
              <Route path="callback" element={<Callback />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </ThemeProvider>
    </>
  );
}
