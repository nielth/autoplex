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
} from "@mui/material";
import { useSearchParams } from "react-router-dom";
import { visuallyHidden } from "@mui/utils";

import DownloadIcon from "@mui/icons-material/Download";
import AccessTimeIcon from "@mui/icons-material/AccessTime";
import DescriptionIcon from "@mui/icons-material/Description";
import ArrowDropUpOutlinedIcon from "@mui/icons-material/ArrowDropUpOutlined";
import ArrowDropDownOutlinedIcon from "@mui/icons-material/ArrowDropDownOutlined";
import CloseIcon from "@mui/icons-material/Close";

const DOMAIN = process.env.REACT_APP_FLASK_LOCATION;


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
}

type Order = "asc" | "desc";

interface HeadCell {
  id: keyof Data;
  label: React.ReactNode;
  numeric: boolean;
  align: any;
  sort: boolean;
  width?: string;
}

const headCells: readonly HeadCell[] = [
  {
    id: "name",
    numeric: false,
    align: "left",
    sort: false,
    label: "Torrent Name",
    width: "930px"
  },
  { id: "download", numeric: true, align: "center", sort: false, label: "" },
  {
    id: "addedTimestamp",
    numeric: false,
    align: "center",
    sort: false,
    label: <AccessTimeIcon fontSize="small" />,
  },
  {
    id: "size",
    numeric: true,
    align: "center",
    sort: false,
    label: <DescriptionIcon fontSize="small" />,
  },
  {
    id: "completed",
    numeric: true,
    align: "center",
    sort: false,
    label: <DownloadIcon fontSize="small" />,
  },
  {
    id: "seeders",
    numeric: true,
    align: "center",
    sort: false,
    label: <ArrowDropUpOutlinedIcon />,
  },
  {
    id: "leechers",
    numeric: true,
    align: "center",
    sort: false,
    label: <ArrowDropDownOutlinedIcon />,
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

function EnhancedTable({ values }: any) {
  const [order, setOrder] = useState<Order>("desc");
  const [orderBy, setOrderBy] = useState<keyof Data>("seeders");
  const rowsPerPage = 35;

  let rows: Data[] = [];

  if (values.searchResp !== undefined) {
    rows = values.searchResp.torrentList.map((torrent: any) => ({
      download: "",
      fid: torrent.fid,
      filename: torrent.filename,
      name: torrent.name,
      addedTimestamp: torrent.addedTimestamp,
      size: torrent.size,
      completed: torrent.completed,
      seeders: torrent.seeders,
      leechers: torrent.leechers,
    }));
  }

  const addSearchParameter = (newParamKey: any, newParamValue: any) => {
    // Create a copy of the current search parameters
    const params = new URLSearchParams(values.searchParams);

    // Add or update the search parameter
    params.set(newParamKey, newParamValue);

    // Set the modified search parameters back to the URL
    values.setSearchParams(params);
  };

  const handleChangeRowsPerPage = (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    values.setPage(0);
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
    values.setPage(newPage);
  };

  let paginatedData;
  if (rows) {
    paginatedData = rows.slice(0, 0 + rowsPerPage);
  }

  return (
    <>
      <Box sx={{ width: "100%" }}>
        <TableContainer>
          <Table sx={{ minWidth: 750 }} aria-labelledby="tableTitle">
            <EnhancedTableHead
              order={order}
              orderBy={orderBy}
              onRequestSort={handleRequestSort}
              rowCount={rows.length}
            />

            {values.searchResp !== undefined ? (
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
                          {row.name}
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
                              axios
                                .post(
                                  `/api/download`,
                                  { fid: row.fid, filename: row.filename },
                                  config
                                )
                                .then(() => {
                                  values.setOpen(true);
                                  setTimeout(() => {
                                    values.setOpen(false);
                                  }, 5000);
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
          count={values.searchResp ? values.searchResp.numFound : 0}
          rowsPerPage={rowsPerPage}
          page={values.page ? values.page : 0}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </Box>
    </>
  );
}

export default function Home() {
  const [searchTerm, setSearchTerm] = useState("");
  const [finalSearch, setFinalSearch] = useState("");
  const [searchResp, setSearchResp] = useState<object>();
  const [open, setOpen] = useState(false);
  const [page, setPage] = useState(-999);

  const [searchParams, setSearchParams] = useSearchParams();

  function search() {
    const config = {
      withCredentials: true,
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-TOKEN": getCookie("csrf_access_token"),
      },
    };
    axios
      .get(
        `/api/search/${searchParams.get("search")}/${
          Number(searchParams.get("p")) + 1
        }`,
        config
      )
      .then((search) => {
        setSearchResp(search.data);
      });
  }

  const handleChange = (event: any) => {
    setSearchTerm(event.target.value);
  };

  const handleSearch = (event: any) => {
    if (event.key === "Enter") {
      setPage(0);
      setFinalSearch(event.target.value);
      setSearchParams(`search=${event.target.value}&p=0`);
    }
  };

  useEffect(() => {
    const query = searchParams.get("search");
    const pa = searchParams.get("p");
    if (query) {
      setSearchTerm(query);
      setFinalSearch(query as string);
    }
    if (pa !== undefined) {
      setPage(Number(pa));
    }
  }, []);

  useEffect(() => {
    if (finalSearch !== "") {
      search();
    }
  }, [finalSearch, page]);

  const values = {
    searchParams,
    setSearchParams,
    open,
    setOpen,
    page,
    setPage,
    searchResp,
  };

  return (
    <>
      <Collapse
        in={open}
        sx={{
          position: "absolute",
          width: "400px",
          left: "0",
          right: "0",
          top: "10px",
        }}
      >
        <Alert
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
          Successfully Added!
        </Alert>
      </Collapse>
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
      <EnhancedTable values={values} />
    </>
  );
}
