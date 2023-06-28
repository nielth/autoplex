import { Box, Button, Table, colors } from "@mui/material";
import { v4 as uuidv4 } from "uuid";
import axios from "axios";
import qs from "qs";
import { useEffect, useState } from "react";
import { PAYLOAD } from "../components/payload";
import "../styles.css";
import { setWithExpiry, getWithExpiry } from "../components/localStorExpire";
import {
  TableContainer,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
} from "@mui/material";
import Paper from "@mui/material/Paper";

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

async function oauthPlexLink() {
  const uuid4: string = uuidv4();
  var url_auth: string = "https://app.plex.tv/auth#!?";
  var forwardUrl: string = "http://localhost:3000/callback";
  var payload = PAYLOAD;
  payload["X-Plex-Client-Identifier"] = uuid4;
  var url_auth_para: string = "";

  await axios
    .post(
      "https://plex.tv/api/v2/pins.json?strong=true",
      qs.stringify(PAYLOAD),
      { headers: {} }
    )
    .then((res) => {
      const code = res.data.code;
      localStorage.setItem("items", JSON.stringify(res.data));
      url_auth_para =
        url_auth +
        "clientID=" +
        uuid4 +
        "&code=" +
        code +
        "&forwardUrl=" +
        forwardUrl;

      setWithExpiry("url", url_auth_para, res.data.expiresIn * 1000);

      console.log(url_auth_para);
    })
    .catch(() => {
      return false;
    });

  return url_auth_para;
}

export function PageA() {
  const [url, setUrl] = useState("");
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    let funcLoggedIn = async () => {
      await axios
        .get("http://192.168.0.165:5000/protected", {
          withCredentials: true,
        })
        .then((resp) => {
          // JSON.parse()
          if (resp.data.logged_in_as) {
            setIsLoggedIn(true);
          } 
        });
        await getWithExpiry("url").then((resp) => {
          if (resp) {
            setUrl(resp);
          } else {
            oauthPlexLink().then((pLink) => {
              setUrl(pLink);
            });
          }
        });
    };

    funcLoggedIn().then(() => {
      setLoading(true)
    })

  }, []);

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
