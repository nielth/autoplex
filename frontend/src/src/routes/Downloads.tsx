import axios from "axios";
import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { authProvider } from "../auth";
import { formatBytes } from "../scripts/formatBytes";
import { getApiDomain } from "../scripts/getApiDomain";

interface DownloadRecord {
  id: number;
  username: string;
  fid: string;
  filename: string;
  torrentSize: number;
  isFreeleech: boolean;
  qbtState?: string;
  progressPercent: number;
  createdAt: string;
  deletedAt?: string;
  deletedByUsername?: string;
  hasPendingDeleteRequest: boolean;
}

interface DeleteRequestRecord {
  id: number;
  downloadEventID: number;
  requestedByUsername: string;
  status: string;
  reason: string;
  approvedByUsername?: string;
  createdAt: string;
  approvedAt?: string;
}

function formatDate(value?: string) {
  if (!value) {
    return "-";
  }

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return "-";
  }

  const year = parsed.getFullYear();
  const month = String(parsed.getMonth() + 1).padStart(2, "0");
  const day = String(parsed.getDate()).padStart(2, "0");
  const hours = String(parsed.getHours()).padStart(2, "0");
  const minutes = String(parsed.getMinutes()).padStart(2, "0");

  return `${year}-${month}-${day} ${hours}:${minutes}`;
}

function normalizeState(state?: string) {
  if (!state) {
    return "unknown";
  }
  return state.replace(/_/g, " ");
}

function toSafeTime(value?: string): number {
  const parsed = Date.parse(value || "");
  return Number.isNaN(parsed) ? 0 : parsed;
}

type DownloadSortField = "createdAt" | "filename" | "username" | "torrentSize";
type DownloadSortDirection = "asc" | "desc";

export function Downloads() {
  const [downloads, setDownloads] = useState<DownloadRecord[]>([]);
  const [pendingRequests, setPendingRequests] = useState<DeleteRequestRecord[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [workingId, setWorkingId] = useState<number | null>(null);
  const [isScanningPlex, setIsScanningPlex] = useState<boolean>(false);
  const [message, setMessage] = useState<string>("");
  const [errorMessage, setErrorMessage] = useState<string>("");
  const [searchTerm, setSearchTerm] = useState<string>("");
  const [selectedUser, setSelectedUser] = useState<string>("all");
  const [sortField, setSortField] = useState<DownloadSortField>("createdAt");
  const [sortDirection, setSortDirection] = useState<DownloadSortDirection>("desc");
  const navigate = useNavigate();
  const domain = getApiDomain();
  const isAdmin = authProvider.isAdmin;

  const availableUsers = useMemo(() => {
    const usersByKey = new Map<string, string>();
    downloads.forEach((download) => {
      const label = (download.username || "").trim();
      if (!label) {
        return;
      }
      const value = label.toLowerCase();
      if (!usersByKey.has(value)) {
        usersByKey.set(value, label);
      }
    });

    return Array.from(usersByKey.entries())
      .map(([value, label]) => ({ value, label }))
      .sort((left, right) =>
        left.label.localeCompare(right.label, undefined, { sensitivity: "base" })
      );
  }, [downloads]);
  const hasMultipleUsers = availableUsers.length > 1;
  const canSortByUser = isAdmin || hasMultipleUsers;

  const loadDownloads = async () => {
    const response = await axios.get(`${domain}/api/downloads`, {
      withCredentials: true,
    });
    setDownloads(response.data.downloads ?? []);
  };

  const loadPendingRequests = async () => {
    if (!isAdmin) {
      setPendingRequests([]);
      return;
    }

    const response = await axios.get(`${domain}/api/downloads/delete-requests`, {
      withCredentials: true,
    });
    setPendingRequests(response.data.requests ?? []);
  };

  const loadAllData = async () => {
    setLoading(true);
    try {
      await Promise.all([loadDownloads(), loadPendingRequests()]);
      setErrorMessage("");
    } catch (error: any) {
      if (error.response?.status === 401) {
        await authProvider.signout();
        navigate("/login");
        return;
      }
      setErrorMessage(error.response?.data?.error || "Failed to load downloads");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAllData();
  }, [domain, isAdmin]);

  useEffect(() => {
    if (!canSortByUser && sortField === "username") {
      setSortField("createdAt");
    }
  }, [canSortByUser, sortField]);

  useEffect(() => {
    if (selectedUser === "all") {
      return;
    }

    const userExists = availableUsers.some((user) => user.value === selectedUser);
    if (!userExists) {
      setSelectedUser("all");
    }
  }, [availableUsers, selectedUser]);

  const handleDelete = async (download: DownloadRecord) => {
    setWorkingId(download.id);
    setMessage("");
    setErrorMessage("");

    try {
      const response = await axios.post(
        `${domain}/api/downloads/${download.id}/delete`,
        {},
        { withCredentials: true }
      );

      if (response.status === 202) {
        setMessage("Delete request submitted");
      } else {
        setMessage("Torrent deleted");
      }

      await Promise.all([loadDownloads(), loadPendingRequests()]);
    } catch (error: any) {
      if (error.response?.status === 401) {
        await authProvider.signout();
        navigate("/login");
        return;
      }

      setErrorMessage(error.response?.data?.error || "Delete action failed");
    } finally {
      setWorkingId(null);
    }
  };

  const handleApprove = async (requestID: number) => {
    setWorkingId(requestID);
    setMessage("");
    setErrorMessage("");

    try {
      await axios.post(
        `${domain}/api/downloads/delete-requests/${requestID}/approve`,
        {},
        { withCredentials: true }
      );
      setMessage("Delete request approved");
      await Promise.all([loadDownloads(), loadPendingRequests()]);
    } catch (error: any) {
      if (error.response?.status === 401) {
        await authProvider.signout();
        navigate("/login");
        return;
      }

      setErrorMessage(error.response?.data?.error || "Failed to approve request");
    } finally {
      setWorkingId(null);
    }
  };

  const handlePlexScan = async () => {
    setIsScanningPlex(true);
    setMessage("");
    setErrorMessage("");

    try {
      const response = await axios.post(
        `${domain}/api/plex/scan/movies-tv`,
        {},
        { withCredentials: true }
      );

      const scannedSections = response.data?.sections;
      if (Array.isArray(scannedSections) && scannedSections.length > 0) {
        setMessage(`Plex scan started for: ${scannedSections.join(", ")}`);
      } else {
        setMessage("Plex scan started for Movies and TV Shows");
      }
    } catch (error: any) {
      if (error.response?.status === 401) {
        await authProvider.signout();
        navigate("/login");
        return;
      }

      setErrorMessage(error.response?.data?.error || "Failed to trigger Plex scan");
    } finally {
      setIsScanningPlex(false);
    }
  };

  const visibleDownloads = useMemo(() => {
    const normalizedSearch = searchTerm.trim().toLowerCase();

    const filtered = downloads.filter((download) => {
      const username = (download.username || "").trim().toLowerCase();
      const matchesUser = selectedUser === "all" || username === selectedUser;
      if (!matchesUser) {
        return false;
      }

      if (!normalizedSearch) {
        return true;
      }

      const title = (download.filename || "").toLowerCase();
      const fid = (download.fid || "").toLowerCase();

      if (title.includes(normalizedSearch) || fid.includes(normalizedSearch)) {
        return true;
      }

      return false;
    });

    const sorted = [...filtered].sort((left, right) => {
      let compare = 0;

      if (sortField === "createdAt") {
        compare = toSafeTime(left.createdAt) - toSafeTime(right.createdAt);
      } else if (sortField === "torrentSize") {
        compare = (left.torrentSize || 0) - (right.torrentSize || 0);
      } else if (sortField === "filename") {
        compare = (left.filename || left.fid || "").localeCompare(
          right.filename || right.fid || "",
          undefined,
          { numeric: true, sensitivity: "base" }
        );
      } else {
        compare = (left.username || "").localeCompare(right.username || "", undefined, {
          numeric: true,
          sensitivity: "base",
        });
      }

      if (compare === 0) {
        compare = left.id - right.id;
      }

      return sortDirection === "asc" ? compare : -compare;
    });

    return sorted;
  }, [downloads, searchTerm, selectedUser, sortDirection, sortField]);

  const handleSortFieldChange = (value: DownloadSortField) => {
    setSortField(value);
    if (value === "createdAt" || value === "torrentSize") {
      setSortDirection("desc");
      return;
    }
    setSortDirection("asc");
  };

  if (loading) {
    return (
      <div className="space-y-3">
        {[...Array(8).keys()].map((index) => (
          <div key={index} className="skeleton h-24 w-full"></div>
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Installed Torrents</h1>
          <p className="text-sm opacity-70">
            {isAdmin
              ? "All tracked torrents across users"
              : "Your tracked torrents and delete options"}
          </p>
        </div>
        <button
          className="btn btn-primary btn-sm"
          disabled={isScanningPlex}
          onClick={handlePlexScan}
        >
          {isScanningPlex ? "Starting Scan..." : "Scan Plex: Movies + TV Shows"}
        </button>
      </div>

      {message ? <div className="alert alert-success">{message}</div> : null}
      {errorMessage ? <div className="alert alert-error">{errorMessage}</div> : null}

      <div className="rounded-xl border border-base-300 bg-base-200 p-4">
        <div className="flex flex-wrap items-end gap-3">
          <label className="form-control w-full sm:max-w-md">
            <span className="label-text text-xs uppercase tracking-wide opacity-70">Search</span>
            <input
              type="text"
              className="input input-bordered input-sm"
              placeholder="Search title or fid"
              value={searchTerm}
              onChange={(event) => setSearchTerm(event.target.value)}
            />
          </label>

          {canSortByUser ? (
            <label className="form-control w-full sm:w-56">
              <span className="label-text text-xs uppercase tracking-wide opacity-70">
                Filter user
              </span>
              <select
                className="select select-bordered select-sm"
                value={selectedUser}
                onChange={(event) => setSelectedUser(event.target.value)}
              >
                <option value="all">All users</option>
                {availableUsers.map((user) => (
                  <option key={user.value} value={user.value}>
                    {user.label}
                  </option>
                ))}
              </select>
            </label>
          ) : null}

          <label className="form-control w-full sm:w-56">
            <span className="label-text text-xs uppercase tracking-wide opacity-70">Sort by</span>
            <select
              className="select select-bordered select-sm"
              value={sortField}
              onChange={(event) =>
                handleSortFieldChange(event.target.value as DownloadSortField)
              }
            >
              <option value="createdAt">Added date</option>
              <option value="torrentSize">Size</option>
              <option value="filename">Title</option>
              {canSortByUser ? <option value="username">User</option> : null}
            </select>
          </label>

          <button
            className="btn btn-outline btn-sm"
            onClick={() => setSortDirection(sortDirection === "asc" ? "desc" : "asc")}
          >
            {sortDirection === "asc" ? "Ascending" : "Descending"}
          </button>
        </div>

        <p className="mt-2 text-xs opacity-70">
          Showing {visibleDownloads.length} of {downloads.length} torrents
        </p>
      </div>

      {downloads.length === 0 ? (
        <div className="rounded-xl border border-base-300 bg-base-200 p-5 text-sm opacity-80">
          No tracked torrents found.
        </div>
      ) : visibleDownloads.length === 0 ? (
        <div className="rounded-xl border border-base-300 bg-base-200 p-5 text-sm opacity-80">
          No torrents match your search.
        </div>
      ) : (
        <div className="space-y-3">
          {visibleDownloads.map((download) => {
            const progress = Math.max(0, Math.min(100, download.progressPercent || 0));
            const deleteLabel =
              isAdmin || download.isFreeleech ? "Delete Torrent" : "Request Delete";

            return (
              <div
                key={download.id}
                className="rounded-xl border border-base-300 bg-base-200 p-3"
              >
                <div className="mb-1 flex flex-wrap items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="font-semibold break-all">{download.filename || download.fid}</p>
                    <p className="text-xs opacity-70">
                      Added: {formatDate(download.createdAt)} - Size:{" "}
                      {formatBytes(download.torrentSize || 0)} - State:{" "}
                      {normalizeState(download.qbtState)}
                      {canSortByUser ? ` - User: ${download.username}` : ""}
                    </p>
                  </div>
                  <div className="flex gap-2">
                    {download.isFreeleech ? (
                      <span className="badge badge-warning badge-sm">FREELEECH</span>
                    ) : null}
                    {download.hasPendingDeleteRequest && !download.deletedAt ? (
                      <span className="badge badge-info badge-sm">Delete Pending</span>
                    ) : null}
                    {download.deletedAt ? (
                      <span className="badge badge-neutral badge-sm">Deleted</span>
                    ) : null}
                  </div>
                </div>

                <div className="mt-1 flex flex-wrap items-center gap-2">
                  <div className="flex min-w-[180px] flex-1 items-center gap-2">
                    <progress
                      className="progress progress-info h-2 flex-1"
                      value={progress}
                      max={100}
                    />
                    <span className="w-14 text-right text-xs tabular-nums opacity-75">
                      {progress.toFixed(2)}%
                    </span>
                  </div>
                  {download.deletedAt ? (
                    <span className="text-xs opacity-70">
                      Deleted by {download.deletedByUsername || "unknown"} at{" "}
                      {formatDate(download.deletedAt)}
                    </span>
                  ) : (
                    <button
                      className="btn btn-error btn-xs sm:btn-sm"
                      disabled={workingId === download.id}
                      onClick={() => handleDelete(download)}
                    >
                      {deleteLabel}
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {isAdmin ? (
        <div className="space-y-3">
          <h2 className="text-xl font-semibold">Pending Delete Requests</h2>
          {pendingRequests.length === 0 ? (
            <div className="rounded-xl border border-base-300 bg-base-200 p-5 text-sm opacity-80">
              No pending delete requests.
            </div>
          ) : (
            pendingRequests.map((request) => (
              <div
                key={request.id}
                className="rounded-xl border border-base-300 bg-base-200 p-4"
              >
                <p className="font-medium">
                  Request #{request.id} for Download #{request.downloadEventID}
                </p>
                <p className="text-xs opacity-70">
                  Requested by {request.requestedByUsername} at{" "}
                  {formatDate(request.createdAt)}
                </p>
                {request.reason ? (
                  <p className="mt-1 text-sm">Reason: {request.reason}</p>
                ) : null}
                <button
                  className="btn btn-success btn-sm mt-3"
                  disabled={workingId === request.id}
                  onClick={() => handleApprove(request.id)}
                >
                  Approve and Delete
                </button>
              </div>
            ))
          )}
        </div>
      ) : null}
    </div>
  );
}
