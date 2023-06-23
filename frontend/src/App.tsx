import "./styles.css";
import NavBar from "./NavBar"

import { createTheme, ThemeProvider } from '@mui/material/styles';
import { PageContainer } from "./PageContainer";

const theme = createTheme({
  palette: {
    background: {
      paper: '#fff',
    },
    text: {
      primary: '#173A5E',
      secondary: '#46505A',
    },
    action: {
      active: '#001E3C',
    },
  },
});


function App() {
  return (
    <>
    <ThemeProvider theme={theme}>
      <NavBar />
      <PageContainer />
    </ThemeProvider>
    </>
  );
}

export default App;
