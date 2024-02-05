import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../App";
import axios from "axios";

export default function LoginPage() {
  const [error, setError] = useState<string>();

  const navigate = useNavigate();
  const location = useLocation();
  const auth = useAuth();

  let from = location.state?.from?.pathname || "/";

  axios.get("/api/authToken")

  return <>Not logged in</>;
}
