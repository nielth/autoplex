import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import axios from "axios";
import TextField from "@mui/material/TextField";
import { useEffect, useState } from "react";
import { getCookie } from "../scripts/getCookie";
import { formatBytes } from "../scripts/formatBytes";
import { Box, Button, Container } from "@mui/material";

import DownloadIcon from "@mui/icons-material/Download";
import AccessTimeIcon from "@mui/icons-material/AccessTime";
import DescriptionIcon from "@mui/icons-material/Description";
import ArrowDropUpOutlinedIcon from "@mui/icons-material/ArrowDropUpOutlined";
import ArrowDropDownOutlinedIcon from "@mui/icons-material/ArrowDropDownOutlined";
import { useSearchParams } from "react-router-dom";

const TorrentTable = ({ torrentList }: any) => {
  return (
    <TableContainer>
      <Table sx={{ minWidth: 650 }} aria-label="simple table">
        <TableHead>
          <TableRow>
            <TableCell>Torrent Name</TableCell>
            <TableCell align="center"></TableCell>
            <TableCell width={"170px"} align="center">
              <Box display={"flex"} justifyContent={"center"}>
                <AccessTimeIcon fontSize="small" />
              </Box>
            </TableCell>
            <TableCell width={"110px"} align="center">
              <Box display={"flex"} justifyContent={"center"}>
                <DescriptionIcon fontSize="small" />
              </Box>
            </TableCell>
            <TableCell align="center">
              <Box display={"flex"} justifyContent={"center"}>
                <DownloadIcon fontSize="small" />
              </Box>
            </TableCell>
            <TableCell align="center">
              <Box display={"flex"} justifyContent={"center"}>
                <ArrowDropUpOutlinedIcon />
              </Box>
            </TableCell>
            <TableCell align="center">
              <Box display={"flex"} justifyContent={"center"}>
                <ArrowDropDownOutlinedIcon />
              </Box>
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {torrentList.torrentList.map((torrent: any) => (
            <TableRow
              key={torrent.fid}
              sx={{ "&:last-child td, &:last-child th": { border: 0 } }}
            >
              <TableCell component="th" scope="row" sx={{fontSize: '16px', fontWeight: 'bold'}}>
                {torrent.name}
              </TableCell>
              <TableCell align="center">
                <Button
                  onClick={() => {
                    const config = {
                      withCredentials: true,
                      headers: {
                        "Content-Type": "application/json",
                        "X-CSRF-TOKEN": getCookie("csrf_access_token"),
                      },
                    };
                    axios
                      .post(
                        `/api/download`,
                        { fid: torrent.fid, filename: torrent.filename },
                        config
                      )
                      .then((search) => {
                        console.log(search);
                      });
                  }}
                >
                  <DownloadIcon />
                </Button>
              </TableCell>
              <TableCell sx={{color: '#978f8f'}} align="center">{torrent.addedTimestamp}</TableCell>
              <TableCell sx={{color: '#978f8f'}} align="center">{formatBytes(torrent.size)}</TableCell>
              <TableCell sx={{color: '#978f8f'}} align="center">{torrent.completed}</TableCell>
              <TableCell sx={{color: '#978f8f'}} align="center">{torrent.seeders}</TableCell>
              <TableCell sx={{color: '#978f8f'}} align="center">{torrent.leechers}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default function Home() {
  const [searchTerm, setSearchTerm] = useState("");
  const [searchResp, setSearchResp] = useState("");

  const [searchParams, setSearchParams] = useSearchParams();

  function search() {
    const config = {
      withCredentials: true,
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-TOKEN": getCookie("csrf_access_token"),
      },
    };
    axios.post("/api/search", { search: searchTerm }, config).then((search) => {
      console.log(search);
      setSearchResp(search.data);
    });
  }

  const handleSearch = (event: any) => {
    if (event.key === "Enter") {
      setSearchParams({ search: searchTerm });
      search();
    }
  };

  const handleChange = (event: any) => {
    setSearchTerm(event.target.value);
  };

  useEffect(() => {
    const query = searchParams.get("search");
    if (query) {
      setSearchTerm(query);
      search();
    }
  }, []);

  return (
    <>
      <Box component={Container} textAlign={"center"} py={5}>
        <TextField
          label="Search"
          variant="outlined"
          value={searchTerm}
          onChange={handleChange}
          onKeyDown={handleSearch}
          color="primary"
          sx={{ width: "50%" }}
        />
      </Box>
      {searchResp !== "" ? <TorrentTable torrentList={searchResp} /> : null}
    </>
  );
}
