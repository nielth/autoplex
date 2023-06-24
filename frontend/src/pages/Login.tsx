import Button from "@mui/material/Button";
import PageContainer from "../PageContainer";

function Login() {
  return (
    <Button
      variant="outlined"
      href="#outlined-buttons"
      size="large"
      color="plex_col"
    >
      Continue with Plex
    </Button>
  );
}

export function LoginPage() {
    return(
        <PageContainer main={<Login />} />
    )
}