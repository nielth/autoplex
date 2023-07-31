import React from "react";
import ReactDOM from "react-dom/client";
import reportWebVitals from "./reportWebVitals";
import { CssBaseline, ThemeProvider, createTheme } from "@mui/material";
import { App } from "./App";

declare module "@mui/material/styles" {
  interface Palette {
    plex: Palette["primary"];
    navbar: Palette["primary"];
    border: Palette["primary"];
  }

  interface PaletteOptions {
    plex?: PaletteOptions["primary"];
    navbar?: PaletteOptions["primary"];
    border?: PaletteOptions["primary"];
  }
}

const theme = createTheme({
  components: {
    MuiButton: {
      defaultProps: {
        disableRipple: true,
        
      },
      styleOverrides: {
        colorInherit: true,
      },
      
    },
    MuiButtonBase: {
      defaultProps: {
        color: 'white'
      }
    },
  },
  palette: {
    background: {
      default: "#0e1116",
    },
    primary: {
      main: "#e7edf2"
    },
    text: {
      primary: "#e7edf2",
    },
    plex: {
      main: "#c17f34",
    },
    navbar: {
      main: '#020409'
    },
    border: {
      main: '#30363d'
    }
  },
  typography: {
    button: {
      fontWeight: "bold",
      textTransform: 'capitalize',
    },
  },
});

const root = ReactDOM.createRoot(
  document.getElementById("root") as HTMLElement
);
root.render(
  <React.StrictMode>
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <App />
    </ThemeProvider>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals(console.log);
