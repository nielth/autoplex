import { Outlet, redirect } from "react-router-dom";
import axios from "axios";
import type { LoaderFunctionArgs } from "react-router-dom";

import { Navbar } from "../components/Navbar";
import { authProvider } from "../auth";

export async function rootLoader({ request }: LoaderFunctionArgs) {
  const resp = await axios
    .get("https://autoplex.nielth.com/api/protected", {
      withCredentials: true,
    })
    .then((resp: any) => {
      console.log(resp.status);
      if (resp.status === 200 && resp.data.logged_in_as) {
        authProvider.signin(resp.data.logged_in_as);
      }
    })
    .catch((err: any) => {
      if (err.response && err.response.status === 401) {
        authProvider.signout();
      }
    });
  if (!authProvider.isAuthenticated) {
    let params = new URLSearchParams();
    params.set("from", new URL(request.url).pathname);
    console.log(true);
    return redirect("/login");
  }
  return null;
}

export default function Root() {
  return (
    <>
      <Navbar />
      <div id="detail" className="pt-8 container mx-auto">
        <Outlet />
      </div>
    </>
  );
}
