import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { createBrowserRouter, RouterProvider } from "react-router-dom";

import "./main.css";

import ErrorPage from "./error-page.tsx";
import { Login } from "./routes/Login.tsx";
import Callback from "./routes/Callback.tsx";
import { Home } from "./routes/Home.tsx";
import Root from "./routes/root.tsx";
import { TvMaze } from "./routes/TvMaze.tsx";
import { Downloads } from "./routes/Downloads.tsx";

const router = createBrowserRouter([
  {
    path: "/",
    element: <Root />,
    errorElement: <ErrorPage />,
    children: [
      {
        index: true,
        Component: Home,
      },
      {
        path: "/tvmaze",
        Component: TvMaze,
      },
      {
        path: "/downloads",
        Component: Downloads,
      },
    ],
  },
  {
    path: "/login",
    Component: Login,
  },
  {
    path: "/callback",
    Component: Callback,
  },
]);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>
);
