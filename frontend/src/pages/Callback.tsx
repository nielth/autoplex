import axios from "axios";

export default function Callback() {
  axios.get("/api/callback").then((resp) => {
    console.log(resp.data);
  });
  return <>Callback</>;
}
