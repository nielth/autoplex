import { Box } from "@mui/material";

type main = {main: any}

const PageContainer = ({main}: main) => {
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
        {main}
      </Box>
    </div>
  )
}

export default PageContainer