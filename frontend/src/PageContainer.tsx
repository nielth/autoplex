import { Box } from "@mui/material";
import Button from "@mui/material/Button";
import { minWidth } from "@mui/system";

export function PageContainer() {
  return (
    <div className="PageContainer">
      <Box
        height={1}
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          width: "300px",
          backgroundColor: "#171b21",
          borderColor: "#202429",
          borderRadius: '7px',
          borderStyle: 'solid',
          borderWidth: '1px',
          minWidth: '400px',
          minHeight: '400px',
          margin: 'auto',
        }}
      >
        <Button
          variant="outlined"
          href="#outlined-buttons"
          size="large"
          color="plex_col"
        >
          Continue with Plex
        </Button>
      </Box>
    </div>
  );
}
