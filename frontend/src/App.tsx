import { createTheme, ThemeProvider } from "@mui/material/styles";
import { Route, Routes, BrowserRouter } from "react-router-dom";

import { Home } from "./pages/home";
import { Temp } from "./pages/temp";
import { Callback } from "./pages/callback";
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

declare module "@mui/material/Button" {
  interface ButtonPropsColorOverrides {
    plex_col: true;
  }
}

export default function App() {
  return (
    <>
      <ThemeProvider theme={theme}>
        <BrowserRouter>
          <Routes>
            <Route element={<NavBar />}>
              <Route path="/" element={<Home />} />
              <Route path="downloads" element={<Temp />} />
              <Route path="callback" element={<Callback />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </ThemeProvider>
    </>
  );
}
