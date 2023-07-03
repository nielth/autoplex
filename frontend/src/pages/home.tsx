import { useEffect, useState } from "react";

import "../styles.css";
import { setWithExpiry, getWithExpiry } from "../components/localStorExpire";
import { getCookie } from "../components/getCookies";
import { logout } from "../api/auth";
import { PAYLOAD } from "../components/payload";
import { funcLoggedIn } from "../api/auth";
import { oauthPlexLink } from "../api/auth";

import { Box, Button, Table, colors } from "@mui/material";
import Paper from "@mui/material/Paper";
import axios from "axios";
import {
  TableContainer,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
} from "@mui/material";

function notLoggedIn(url: string) {
  return (
    <div>
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
            borderRadius: "7px",
            borderStyle: "solid",
            borderWidth: "1px",
            minWidth: "400px",
            minHeight: "400px",
            margin: "auto",
          }}
        >
          <Button
            variant="outlined"
            onClick={function () {
              localStorage.removeItem("url");
            }}
            href={url}
            size="large"
            color="plex_col"
          >
            Continue with Plex
          </Button>
        </Box>
      </div>
    </div>
  );
}

function createData(
  name: string,
  calories: number,
  fat: number,
  carbs: number,
  protein: number
) {
  return { name, calories, fat, carbs, protein };
}

const rows = [
  createData("Frozen yoghurt", 159, 6.0, 24, 4.0),
  createData("Ice cream sandwich", 237, 9.0, 37, 4.3),
  createData("Eclair", 262, 16.0, 24, 6.0),
  createData("Cupcake", 305, 3.7, 67, 4.3),
  createData("Gingerbread", 356, 16.0, 49, 3.9),
];

function loggedIn() {
  return (
    <>
      <div id="pageContainer">
        <div id="table">
          <TableContainer component={Paper}>
            <Table
              sx={{ minWidth: 650 }}
              size="small"
              aria-label="a dense table"
            >
              <TableHead>
                <TableRow>
                  <TableCell>Dessert (100g serving)</TableCell>
                  <TableCell align="right">Calories</TableCell>
                  <TableCell align="right">Fat&nbsp;(g)</TableCell>
                  <TableCell align="right">Carbs&nbsp;(g)</TableCell>
                  <TableCell align="right">Protein&nbsp;(g)</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {rows.map((row) => (
                  <TableRow
                    key={row.name}
                    sx={{ "&:last-child td, &:last-child th": { border: 0 } }}
                  >
                    <TableCell component="th" scope="row">
                      {row.name}
                    </TableCell>
                    <TableCell align="right">{row.calories}</TableCell>
                    <TableCell align="right">{row.fat}</TableCell>
                    <TableCell align="right">{row.carbs}</TableCell>
                    <TableCell align="right">{row.protein}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </div>
      </div>
    </>
  );
}

async function setUrlFunc() {
  const temp = await getWithExpiry("url");
  if (!temp) {
    oauthPlexLink();
  }
  return temp;
}

export function PageA() {
  const [url, setUrl] = useState("");
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    funcLoggedIn().then((resp) => {
      if (resp.data.logged_in_as) {
        setIsLoggedIn(true);
      }
      setLoading(true);
    });

    setUrlFunc().then((resp: any) => {
      setUrl(resp);
    });
    
  }, [isLoggedIn]);

  return (
    <>
      {loading ? (
        isLoggedIn ? (
          loggedIn()
        ) : url ? (
          notLoggedIn(url)
        ) : (
          <div />
        )
      ) : (
        <div />
      )}
    </>
  );
}
