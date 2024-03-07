import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import axios from "axios";
import TextField from "@mui/material/TextField";
import { useEffect, useState } from "react";
import { getCookie } from "../scripts/getCookie";
import { formatBytes } from "../scripts/formatBytes";
import {
  Alert,
  Box,
  Button,
  Collapse,
  IconButton,
  TablePagination,
  TableSortLabel,
  Tooltip,
  duration,
} from "@mui/material";
import { useSearchParams } from "react-router-dom";
import { visuallyHidden } from "@mui/utils";

import DownloadIcon from "@mui/icons-material/Download";
import AccessTimeIcon from "@mui/icons-material/AccessTime";
import DescriptionIcon from "@mui/icons-material/Description";
import ArrowDropUpOutlinedIcon from "@mui/icons-material/ArrowDropUpOutlined";
import ArrowDropDownOutlinedIcon from "@mui/icons-material/ArrowDropDownOutlined";
import CloseIcon from "@mui/icons-material/Close";
import { useAuth } from "../App";

interface Data {
  download: string;
  fid: string;
  filename: string;
  name: string;
  addedTimestamp: string;
  size: number;
  completed: number;
  seeders: number;
  leechers: number;
  tags: string[];
  categoryID: number;
}

type Order = "asc" | "desc";

interface HeadCell {
  id: keyof Data;
  label: React.ReactNode;
  numeric: boolean;
  align: any;
  sort: boolean;
  width?: string;
  minWidth?: string;
  color?: string;
}

const headCells: readonly HeadCell[] = [
  {
    id: "name",
    numeric: false,
    align: "left",
    sort: false,
    label: "Torrent Name",
    width: "930px",
  },
  { id: "download", numeric: true, align: "center", sort: false, label: "" },
  {
    id: "addedTimestamp",
    numeric: false,
    align: "center",
    sort: false,
    label: (
      <Tooltip title="Added Timestamp">
        <AccessTimeIcon fontSize="small" />
      </Tooltip>
    ),
    minWidth: "180px",
  },
  {
    id: "size",
    numeric: true,
    align: "center",
    sort: false,
    label: (
      <Tooltip title="Size">
        <DescriptionIcon fontSize="small" />
      </Tooltip>
    ),
    minWidth: "105px",
  },
  {
    id: "completed",
    numeric: true,
    align: "center",
    sort: false,
    label: (
      <Tooltip title="Time Downloaded">
        <DownloadIcon fontSize="small" />
      </Tooltip>
    ),
  },
  {
    id: "seeders",
    numeric: true,
    align: "center",
    sort: false,
    label: (
      <Tooltip title="Seeders">
        <ArrowDropUpOutlinedIcon />
      </Tooltip>
    ),
    color: "green",
  },
  {
    id: "leechers",
    numeric: true,
    align: "center",
    sort: false,
    label: (
      <Tooltip title="Leecers">
        <ArrowDropDownOutlinedIcon />
      </Tooltip>
    ),
    color: "red",
  },
];

interface EnhancedTableProps {
  onRequestSort: (
    event: React.MouseEvent<unknown>,
    property: keyof Data
  ) => void;
  order: Order;
  orderBy: string;
  rowCount: number;
}

function CollapseInteract({
  value,
  failed,
}: {
  value: boolean;
  failed: boolean;
}) {
  const [open, setOpen] = useState(value);

  useEffect(() => {
    setOpen(value);

    if (value) {
      const timer = setTimeout(() => {
        setOpen(false);
      }, 5000);

      // Clear the timer if the component unmounts or if the value changes
      return () => clearTimeout(timer);
    }
  }, [value]);

  return (
    <Collapse
      in={open}
      sx={{
        margin: "auto",
        position: "fixed",
        width: "400px",
        left: "0",
        right: "0",
        top: "10px",
      }}
    >
      <Alert
        severity={!failed ? "success" : "error"}
        action={
          <IconButton
            aria-label="close"
            color="inherit"
            size="small"
            onClick={() => {
              setOpen(false);
            }}
          >
            <CloseIcon fontSize="inherit" />
          </IconButton>
        }
        sx={{ mb: 2 }}
      >
        {!failed ? (
          <>Successfully Added!</>
        ) : (
          <>Could Not Start Download, contant admin!</>
        )}
      </Alert>
    </Collapse>
  );
}

function EnhancedTableHead(props: EnhancedTableProps) {
  const { order, orderBy, onRequestSort } = props;
  const createSortHandler =
    (property: keyof Data) => (event: React.MouseEvent<unknown>) => {
      onRequestSort(event, property);
    };

  return (
    <TableHead>
      <TableRow>
        {headCells.map((headCell) => (
          <TableCell
            key={headCell.id}
            align={headCell.align}
            sortDirection={orderBy === headCell.id ? order : false}
            width={headCell.width}
            sx={{ minWidth: headCell.minWidth, color: headCell.color }}
          >
            {headCell.sort ? (
              <TableSortLabel
                active={orderBy === headCell.id}
                direction={orderBy === headCell.id ? order : "asc"}
                onClick={createSortHandler(headCell.id)}
              >
                {headCell.label}
                {orderBy === headCell.id ? (
                  <Box component="span" sx={visuallyHidden}>
                    {order === "desc"
                      ? "sorted descending"
                      : "sorted ascending"}
                  </Box>
                ) : null}
              </TableSortLabel>
            ) : (
              <>
                {headCell.label}
                {orderBy === headCell.id ? (
                  <Box component="span" sx={visuallyHidden}>
                    {order === "desc"
                      ? "sorted descending"
                      : "sorted ascending"}
                  </Box>
                ) : null}
              </>
            )}
          </TableCell>
        ))}
      </TableRow>
    </TableHead>
  );
}

interface Test {
  torrentList: Object[];
  numFound: number;
}

function search(searchParams: any) {
  const DOMAIN = process.env.REACT_APP_FLASK_LOCATION || "";

  const config = {
    withCredentials: true,
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-TOKEN": getCookie("csrf_access_token"),
    },
  };
  return axios.get(
    `${DOMAIN}/api/search/${searchParams.get("search")}/${
      Number(searchParams.get("p")) + 1
    }`,
    config
  );
}

export default function Home() {
  const [finalSearch, setFinalSearch] = useState("");
  const [searchResp, setSearchResp] = useState<Test>();
  const [page, setPage] = useState(0);
  const [failedDownload, setFailedDownload] = useState<boolean>(false);
  const [_, setSearchTerm] = useState("");
  const [order, setOrder] = useState<Order>("desc");
  const [orderBy, setOrderBy] = useState<keyof Data>("seeders");
  const [CollapseInteractOpen, CollapseInteractSetOpen] = useState(false);
  const [searchParams, setSearchParams] = useSearchParams();
  let auth = useAuth();
  const rowsPerPage = 35;

  const DOMAIN = process.env.REACT_APP_FLASK_LOCATION || "";

  let rows: Data[] = [];

  if (searchResp !== undefined) {
    rows = searchResp.torrentList.map((torrent: any) => ({
      download: "",
      fid: torrent.fid,
      filename: torrent.filename,
      name: torrent.name,
      addedTimestamp: torrent.addedTimestamp,
      size: torrent.size,
      completed: torrent.completed,
      seeders: torrent.seeders,
      leechers: torrent.leechers,
      tags: torrent.tags,
      categoryID: torrent.categoryID,
    }));
  }

  const addSearchParameter = (newParamKey: any, newParamValue: any) => {
    // Create a copy of the current search parameters
    const params = new URLSearchParams(searchParams);

    // Add or update the search parameter
    params.set(newParamKey, newParamValue);

    // Set the modified search parameters back to the URL
    setSearchParams(params);
  };

  const handleChangeRowsPerPage = (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    setPage(0);
  };

  const handleRequestSort = (
    event: React.MouseEvent<unknown>,
    property: keyof Data
  ) => {
    const isAsc = orderBy === property && order === "asc";
    setOrder(isAsc ? "desc" : "asc");
    setOrderBy(property);
  };

  const handleChangePage = (event: unknown, newPage: number) => {
    addSearchParameter("p", newPage);
    setPage(newPage);
    window.scrollTo(0, 0);
  };

  let paginatedData;
  if (rows) {
    paginatedData = rows.slice(0, 0 + rowsPerPage);
  }

  const handleSearch = (event: any) => {
    if (event.key === "Enter") {
      setPage(0);
      setFinalSearch(event.target.value);
      setSearchParams(`search=${event.target.value}&p=0`);
    }
  };

  const handleChange = (event: any) => {
    setSearchTerm(event.target.value);
  };

  useEffect(() => {
    if (finalSearch !== "") {
      search(searchParams).then((search) => {
        setSearchResp(search.data);
      });
    }
  }, [finalSearch, page]);

  useEffect(() => {}, [CollapseInteractOpen]);

  useEffect(() => {
    const query = searchParams.get("search");
    const pa = searchParams.get("p");
    if (query) {
      setFinalSearch(query as string);
    }
    if (pa !== undefined) {
      setPage(Number(pa));
    }
  }, [searchParams.get("search"), searchParams.get("p")]);

  return (
    <>
      <CollapseInteract value={CollapseInteractOpen} failed={failedDownload} />
      <Box>
        <TextField
          label="Search"
          variant="outlined"
          onChange={handleChange}
          onKeyDown={handleSearch}
          color="primary"
          sx={{ width: "50%" }}
        />
      </Box>
      <Box>
        <Box sx={{ width: "100%" }}>
          <TableContainer>
            <Table sx={{ minWidth: 750 }} aria-labelledby="tableTitle">
              <EnhancedTableHead
                order={order}
                orderBy={orderBy}
                onRequestSort={handleRequestSort}
                rowCount={rows.length}
              />

              {searchResp !== undefined ? (
                <TableBody>
                  {paginatedData &&
                    paginatedData.map((row) => {
                      return (
                        <TableRow hover tabIndex={-1} key={Number(row.fid)}>
                          <TableCell
                            component="th"
                            scope="row"
                            sx={{ fontSize: "16px", fontWeight: "bold" }}
                          >
                            <>
                              <Box
                                display={"flex"}
                                gap={1}
                                alignItems={"center"}
                              >
                                <Box>{row.name}</Box>
                                {row.tags.includes("FREELEECH") ? (
                                  <>
                                    <Box>
                                      <Box
                                        bgcolor={"#FFDF00"}
                                        color={"#4d4d4d"}
                                        borderRadius={"2px"}
                                        padding={"2px 4px"}
                                        fontSize={"11px"}
                                      >
                                        FREELEECH
                                      </Box>
                                    </Box>
                                  </>
                                ) : null}
                              </Box>
                            </>
                          </TableCell>
                          <TableCell align="center">
                            <Button
                              onClick={() => {
                                const config = {
                                  withCredentials: true,
                                  headers: {
                                    "Content-Type": "application/json",
                                    "X-CSRF-TOKEN":
                                      getCookie("csrf_access_token"),
                                  },
                                };
                                CollapseInteractSetOpen(false);
                                setFailedDownload(false);
                                axios
                                  .post(
                                    `${DOMAIN}/api/download`,
                                    {
                                      fid: row.fid,
                                      filename: row.filename,
                                      categoryID: row.categoryID,
                                    },
                                    config
                                  )
                                  .then(() => {})
                                  .catch((resp) => {
                                    if (resp.response.status === 401) {
                                      console.log(2);
                                      auth.signout(() => {});
                                    }
                                    setFailedDownload(true);
                                  })
                                  .finally(() => {
                                    CollapseInteractSetOpen(true);
                                  });
                              }}
                            >
                              <DownloadIcon />
                            </Button>
                          </TableCell>
                          <TableCell sx={{ color: "#978f8f" }} align="center">
                            {row.addedTimestamp}
                          </TableCell>
                          <TableCell sx={{ color: "#978f8f" }} align="center">
                            {formatBytes(row.size as number)}
                          </TableCell>
                          <TableCell sx={{ color: "#978f8f" }} align="center">
                            {row.completed}
                          </TableCell>
                          <TableCell sx={{ color: "#978f8f" }} align="center">
                            {row.seeders}
                          </TableCell>
                          <TableCell sx={{ color: "#978f8f" }} align="center">
                            {row.leechers}
                          </TableCell>
                        </TableRow>
                      );
                    })}
                </TableBody>
              ) : null}
            </Table>
          </TableContainer>
          <TablePagination
            rowsPerPageOptions={[]}
            component="div"
            count={searchResp ? searchResp.numFound : 0}
            rowsPerPage={rowsPerPage}
            page={page ? page : 0}
            onPageChange={handleChangePage}
            onRowsPerPageChange={handleChangeRowsPerPage}
          />
        </Box>
      </Box>
    </>
  );
}
