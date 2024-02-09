import {
  Routes,
  Route,
  useLocation,
  Navigate,
  BrowserRouter,
} from "react-router-dom";
import { createContext, useContext, useEffect, useState } from "react";
import axios from "axios";

import Layout from "./pages/Layout";
import LoginPage from "./pages/Login";
import Callback from "./pages/Callback";
import Home from "./pages/Home";

export default function App() {
  return (
    <>
      <BrowserRouter>
        <AuthProvider>
          <Routes>
            <Route
              path="/"
              element={
                <RequireAuth>
                  <Layout />
                </RequireAuth>
              }
            >
              <Route index element={<Home />} />
            </Route>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/callback" element={<Callback />} />
          </Routes>
        </AuthProvider>
      </BrowserRouter>
    </>
  );
}

interface AuthContextType {
  user: any;
  signin: (callback: (success: number) => void) => void;
  signout: (callback: VoidFunction) => void;
}

let AuthContext = createContext<AuthContextType>(null!);

function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        const config = {
          withCredentials: true,
          headers: {
            "Content-Type": "application/json",
          },
        };
        const response = await axios.get("/api/protected", config);
        if (response.status === 200) {
          setUser(response.data.logged_in_as);
        }
      } catch (error) {
        console.error("Error checking auth status", error);
      } finally {
        setIsLoading(false);
      }
    })();
  }, []);

  let signin = (callback: (success: number) => void) => {
    axios
      .get("/api/callback")
      .then((resp: any) => {
        if (resp.status === 200) {
          setUser(resp.data.logged_in_as);
          callback(0);
        }
      })
      .catch((error) => {
        console.log(error.status);
        callback(error.status);
      });
  };

  let signout = (callback: VoidFunction) => {
    const config = {
      withCredentials: true,
      headers: {
        "Content-Type": "application/json",
      },
    };
    axios.get("/api/logout", config).then((resp) => {
      if (resp.status === 200) {
        setUser(null);
        callback();
      }
    });
  };

  let value = { user, signin, signout };

  return (
    <>
      {!isLoading ? (
        <>
          <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
        </>
      ) : null}
    </>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}

function RequireAuth({ children }: { children: JSX.Element }) {
  let auth = useAuth();
  let location = useLocation();

  if (!auth.user) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  return children;
}
