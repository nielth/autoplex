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
import ArrowCircleDownIcon from "@mui/icons-material/ArrowCircleDown";
import { torrentPost, torrentSearch } from "../api/auth";
import { formatBytes } from "../components/formatBytes";
import { useEffect, useState } from "react";

const StyledTableCell = styled(TableCell)(({ theme }) => ({
  [`&.${tableCellClasses.head}`]: {},
  [`&.${tableCellClasses.body}`]: {
    fontSize: 15,
  },
  color: "#e7edf2",
  borderBottom: "1px solid #161b22",
  lineHeight: "21px",
}));

export function NotLoggedIn(context:any) {
  console.log(context.url)
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
            href={context.url}
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

async function torrentDownload() {
  const resp: any = await torrentPost();
  if (resp) {
    return true;
  }
}

function freeleech(tags: any) {
  if (tags.includes("FREELEECH")) {
    return (
      <>
        <span
          style={{
            color: "#4d4d4d",
            backgroundColor: "#FFDF00",
            fontWeight: "500",
            border: "3px solid #FFDF00",
            borderRadius: "2px",
            margin: "0 10px",
          }}
        >
          FREELEECH
        </span>
      </>
    );
  }
}

export function LoggedIn() {
  const [torrentList, setTorrentList] = useState<any>("");
  const handleKeyDown = (event: any) => {
    if (event.key === "Enter") {
      torrentSearch(event.target.value).then((res: any) => {
        setTorrentList(res);
      });
    }
  };

  return (
    <>
      <div className="search-bar">
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
      {torrentList ? (
        <div className="data">
          <div id="table">
            <TableContainer
              component={Paper}
              sx={{
                backgroundColor: "transparent",
                border: "1px solid #30363d",
              }}
            >
              <Table sx={{ minWidth: 700 }} aria-label="customized table">
                <TableHead sx={{ backgroundColor: "#161b22" }}>
                  <TableRow>
                    <StyledTableCell align="center">
                      Torrent Name
                    </StyledTableCell>
                    <StyledTableCell align="center"></StyledTableCell>
                    <StyledTableCell align="center" sx={{ minWidth: "100px;" }}>
                      Size
                    </StyledTableCell>
                    <StyledTableCell align="center">Seeders</StyledTableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {torrentList!.torrentList.map((row: any) => (
                    <TableRow key={row.fid}>
                      <StyledTableCell component="th" scope="row">
                        <div>
                          {row.name}
                          {freeleech(row.tags)}
                        </div>
                        <div style={{ color: "#978f8f", fontSize: "12px" }}>
                          Added: {row.addedTimestamp} | Completed:{" "}
                          {row.completed}
                        </div>
                      </StyledTableCell>
                      <StyledTableCell
                        align="center"
                        component="th"
                        scope="row"
                      >
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
                      <StyledTableCell align="right" component="th" scope="row">
                        {formatBytes(row.size)}
                      </StyledTableCell>
                      <StyledTableCell
                        align="center"
                        component="th"
                        scope="row"
                      >
                        {row.seeders}
                      </StyledTableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </div>
        </div>
      ) : (
        <div />
      )}
    </>
  );
}

export function Home() {
  const context: Dict = useOutletContext();

  return (
    <>
      {context.loading ? (
        context.isLoggedIn ? (
          <LoggedIn />
        ) : context.url ? (
          <NotLoggedIn url={context.url} />
        ) : (
          <div />
        )
      ) : (
        <div />
      )}
    </>
  );
}
