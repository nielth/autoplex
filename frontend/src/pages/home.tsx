import "../styles.css";
import { useOutletContext } from "react-router-dom";

import Paper from "@mui/material/Paper";
import {
  TableContainer,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
  Box,
  Button,
  Table,
} from "@mui/material";
import { Dict } from "styled-components/dist/types";
import { tableCellClasses } from "@mui/material/TableCell";
import { styled } from "@mui/material/styles";
import TextField from "@mui/material/TextField";
import torrentList from "./data.json";
import ArrowCircleDownIcon from "@mui/icons-material/ArrowCircleDown";
import { torrentPost, torrentSearch } from "../api/auth";
import { useState } from "react";

const StyledTableCell = styled(TableCell)(({ theme }) => ({
  [`&.${tableCellClasses.head}`]: {},
  [`&.${tableCellClasses.body}`]: {
    fontSize: 13,
  },
  color: "#e7edf2",
  borderBottom: "1px solid #161b22",
}));

const StyledTableRow = styled(TableRow)(({ theme }) => ({
  textAlign: "center",
  // "&:nth-of-type(odd)": {
  //   backgroundColor: "transparent",
  // },
  "&:last-child td, &:last-child th": {
    border: 0,
  },
}));

const StyledTableHead = styled(TableHead)(({ theme }) => ({}));

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

function formatBytes(bytes: number, decimals = 2) {
  if (!+bytes) return "0 Bytes";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = [
    "Bytes",
    "KiB",
    "MiB",
    "GiB",
    "TiB",
    "PiB",
    "EiB",
    "ZiB",
    "YiB",
  ];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
}

async function torrentDownload() {
  const resp: any = await torrentPost();
  if (resp) {
    return true;
  }
}

function data() {
  return (
    <>
      <div className="data">
        <div id="table">
          <TableContainer
            component={Paper}
            sx={{ backgroundColor: "transparent", border: "1px solid #30363d" }}
          >
            <Table sx={{ minWidth: 700 }} aria-label="customized table">
              <TableHead sx={{ backgroundColor: "#161b22" }}>
                <TableRow>
                  <StyledTableCell align="center">Torrent Name</StyledTableCell>
                  <StyledTableCell align="center"></StyledTableCell>
                  <StyledTableCell align="center">Added</StyledTableCell>
                  <StyledTableCell align="center">Size</StyledTableCell>
                  <StyledTableCell align="center">Completed</StyledTableCell>
                  <StyledTableCell align="center">Seeders</StyledTableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {torrentList.torrentList.map((row) => (
                  <StyledTableRow key={row.fid}>
                    <StyledTableCell component="th" scope="row">
                      {row.name}
                    </StyledTableCell>
                    <StyledTableCell align="center" component="th" scope="row">
                      <Button
                        onClick={torrentDownload}
                        sx={{
                          "&:hover": { backgroundColor: "inherit" },
                          ".MuiTouchRipple-root span": {
                            backgroundColor: "inherit",
                          },
                        }}
                      >
                        <ArrowCircleDownIcon
                          fontSize="medium"
                          sx={{ verticalAlign: "bottom", color: "green" }}
                        />
                      </Button>
                    </StyledTableCell>
                    <StyledTableCell align="center" component="th" scope="row">
                      {row.addedTimestamp}
                    </StyledTableCell>
                    <StyledTableCell align="center" component="th" scope="row">
                      {formatBytes(row.size)}
                    </StyledTableCell>
                    <StyledTableCell align="center" component="th" scope="row">
                      {row.completed}
                    </StyledTableCell>
                    <StyledTableCell align="center" component="th" scope="row">
                      {row.seeders}
                    </StyledTableCell>
                  </StyledTableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </div>
      </div>
    </>
  );
}

function searchBox() {
  const handleKeyDown = (event:any) => {
    if (event.key === 'Enter') {
      torrentSearch(event.target.value)
    }
  };

  return (
    <>
      <div className="serach-bar">
        <div className="main">
          <div className="search">
            <TextField
              id="outlined-basic"
              variant="outlined"
              placeholder="Search movie or show"
              fullWidth
              onKeyDown={handleKeyDown}
              label="Search"
              sx={{
                border: "1px solid #c17f34",
              }}
            />
          </div>
        </div>
      </div>
      {data()}
    </>
  );
}

function loggedIn() {
  return <>{searchBox()}</>;
}

export function Home() {
  const context: Dict = useOutletContext();

  return (
    <>
      {context.loading ? (
        context.isLoggedIn ? (
          loggedIn()
        ) : context.url ? (
          notLoggedIn(context.url)
        ) : (
          <div />
        )
      ) : (
        <div />
      )}
    </>
  );
}
