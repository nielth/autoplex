import { useNavigate } from "react-router-dom";
import axios from "axios";
import { useEffect, useState } from "react";
import { useAuth } from "../App";

export default function Layout() {
  let [content, setContent] = useState<object>();
  let auth = useAuth();
  let navigate = useNavigate();

  useEffect(() => {
    const config: object = {
      withCredentials: true,
      headers: {
        "Content-Type": "application/json",
      },
    };
    axios.get("/api/content", config).then((resp) => {
      setContent(resp.data);
    });
  }, []);

  return <>123</>;
}
