import AppBar from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import Toolbar from "@mui/material/Toolbar";
import Button from "@mui/material/Button";
import Avatar from "@mui/material/Avatar";

var arr = {
  Logs: "/logs",
  Downloads: "/downloads",
};

export default function ButtonAppBar() {
  return (
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
  );
}
