import { Box } from "@mui/material";
import Button from "@mui/material/Button";

export function PageContainer() {
  return (
    <div className="PageContainer">
      <Box
        height={1}
        sx={{ display: "flex", alignItems: "center", justifyContent: "center" }}
      >
        <Button
          variant="outlined"
          href="#outlined-buttons"
          sx={{ width: 1 / 2, color: "rgb(193 127 52)", borderColor: "rgb(193 127 52)", transition:"background-color rgb(153, 99, 37)", forcedColorAdjust: "rgb(153, 99, 37)" }}
        >
          Continue with Plex
        </Button>
      </Box>
    </div>
  );
}
