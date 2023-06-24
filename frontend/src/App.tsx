import "./styles.css";
import NavBar from "./NavBar";
import setToken from "./components/test"
import { LoginPage } from "./pages/Login";

import { createTheme, ThemeProvider } from "@mui/material/styles";

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
        <NavBar />
        <LoginPage />
      </ThemeProvider>
    </>
  );
}

export default App;
