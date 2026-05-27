import axios from "axios";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
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
  hasHitAndRun: boolean;
  completedAt?: string;
  safeToDeleteAt?: string;
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
  safeToDeleteAt?: string;
  autoDeleteAt?: string;
  downloadFilename?: string;
  downloadFid?: string;
  downloadSize?: number;
  downloadIsFreeleech?: boolean;
}

type TabKey = "installed" | "deleted";
type DownloadSortField =
  | "createdAt"
  | "filename"
  | "username"
  | "torrentSize"
  | "deletedAt";
type DownloadSortDirection = "asc" | "desc";

const PAGE_SIZE = 50;

function formatCountdown(targetISO?: string): string {
  if (!targetISO) return "-";
  const target = Date.parse(targetISO);
  if (Number.isNaN(target)) return "-";
  const diffMs = target - Date.now();
  if (diffMs <= 0) return "now";
  const totalMinutes = Math.round(diffMs / 60000);
  const days = Math.floor(totalMinutes / (60 * 24));
  const hours = Math.floor((totalMinutes % (60 * 24)) / 60);
  const minutes = totalMinutes % 60;
  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}

function formatDate(value?: string) {
  if (!value) return "-";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return "-";
  const year = parsed.getFullYear();
  const month = String(parsed.getMonth() + 1).padStart(2, "0");
  const day = String(parsed.getDate()).padStart(2, "0");
  const hours = String(parsed.getHours()).padStart(2, "0");
  const minutes = String(parsed.getMinutes()).padStart(2, "0");
  return `${year}-${month}-${day} ${hours}:${minutes}`;
}

function normalizeState(state?: string) {
  if (!state) return "unknown";
  return state.replace(/_/g, " ");
}

interface TabState {
  rows: DownloadRecord[];
  total: number;
  offset: number;
  loading: boolean;
  initiated: boolean;
}

const emptyTab: TabState = {
  rows: [],
  total: 0,
  offset: 0,
  loading: false,
  initiated: false,
};

export function Downloads() {
  const [activeTab, setActiveTab] = useState<TabKey>("installed");
  const [installed, setInstalled] = useState<TabState>(emptyTab);
  const [deleted, setDeleted] = useState<TabState>(emptyTab);
  const [availableUsers, setAvailableUsers] = useState<string[]>([]);

  const [pendingRequests, setPendingRequests] = useState<DeleteRequestRecord[]>([]);
  const [hitAndRunRequests, setHitAndRunRequests] = useState<DeleteRequestRecord[]>([]);
  const [sideLoading, setSideLoading] = useState<boolean>(true);

  const [workingId, setWorkingId] = useState<number | null>(null);
  const [isScanningPlex, setIsScanningPlex] = useState<boolean>(false);
  const [message, setMessage] = useState<string>("");
  const [errorMessage, setErrorMessage] = useState<string>("");

  const [searchTerm, setSearchTerm] = useState<string>("");
  const [debouncedSearch, setDebouncedSearch] = useState<string>("");
  const [selectedUser, setSelectedUser] = useState<string>("all");
  const [sortField, setSortField] = useState<DownloadSortField>("createdAt");
  const [sortDirection, setSortDirection] = useState<DownloadSortDirection>("desc");

  const navigate = useNavigate();
  const domain = getApiDomain();
  const isAdmin = authProvider.isAdmin;

  const setTabState = (tab: TabKey, updater: (prev: TabState) => TabState) => {
    if (tab === "installed") {
      setInstalled(updater);
    } else {
      setDeleted(updater);
    }
  };

  const requestSeq = useRef<{ installed: number; deleted: number }>({
    installed: 0,
    deleted: 0,
  });

  const loadTab = useCallback(
    async (tab: TabKey, opts: { append: boolean; offset?: number }) => {
      const offset = opts.offset ?? 0;
      const seqId = ++requestSeq.current[tab];
      setTabState(tab, (prev) => ({ ...prev, loading: true, initiated: true }));

      try {
        const response = await axios.get(`${domain}/api/downloads`, {
          withCredentials: true,
          params: {
            status: tab === "installed" ? "active" : "deleted",
            q: debouncedSearch || undefined,
            user:
              isAdmin && selectedUser !== "all" ? selectedUser : undefined,
            sort: sortField,
            dir: sortDirection,
            limit: PAGE_SIZE,
            offset,
          },
        });

        // Stale-response guard: ignore if a newer request started after us.
        if (seqId !== requestSeq.current[tab]) return;

        const incoming: DownloadRecord[] = response.data.downloads ?? [];
        const total: number = response.data.total ?? 0;
        const users: string[] = response.data.availableUsers ?? [];
        if (isAdmin && users.length > 0) {
          setAvailableUsers(users);
        }

        setTabState(tab, (prev) => ({
          rows: opts.append ? [...prev.rows, ...incoming] : incoming,
          total,
          offset: offset + incoming.length,
          loading: false,
          initiated: true,
        }));
        setErrorMessage("");
      } catch (error: any) {
        if (error.response?.status === 401) {
          await authProvider.signout();
          navigate("/login");
          return;
        }
        setErrorMessage(error.response?.data?.error || "Failed to load downloads");
        setTabState(tab, (prev) => ({ ...prev, loading: false }));
      }
    },
    [domain, debouncedSearch, selectedUser, sortField, sortDirection, isAdmin, navigate]
  );

  const loadSideRequests = useCallback(async () => {
    setSideLoading(true);
    try {
      const promises: Promise<unknown>[] = [
        axios
          .get(`${domain}/api/downloads/delete-requests/hit-and-run`, {
            withCredentials: true,
          })
          .then((r) => setHitAndRunRequests(r.data.requests ?? [])),
      ];
      if (isAdmin) {
        promises.push(
          axios
            .get(`${domain}/api/downloads/delete-requests`, {
              withCredentials: true,
            })
            .then((r) => setPendingRequests(r.data.requests ?? []))
        );
      } else {
        setPendingRequests([]);
      }
      await Promise.all(promises);
    } catch (error: any) {
      if (error.response?.status === 401) {
        await authProvider.signout();
        navigate("/login");
        return;
      }
      setErrorMessage(
        error.response?.data?.error || "Failed to load delete queue"
      );
    } finally {
      setSideLoading(false);
    }
  }, [domain, isAdmin, navigate]);

  // Debounce search input — 300ms.
  useEffect(() => {
    const t = window.setTimeout(() => setDebouncedSearch(searchTerm), 300);
    return () => window.clearTimeout(t);
  }, [searchTerm]);

  // First mount: load side panels.
  useEffect(() => {
    loadSideRequests();
  }, [loadSideRequests]);

  // Whenever filters/sort/tab change → reload current tab from offset 0.
  // We also lazily fetch the other tab on first activation only.
  useEffect(() => {
    loadTab(activeTab, { append: false, offset: 0 });
  }, [activeTab, debouncedSearch, selectedUser, sortField, sortDirection, loadTab]);

  const reloadAll = useCallback(async () => {
    await Promise.all([
      loadSideRequests(),
      loadTab("installed", { append: false, offset: 0 }),
      // Only refetch the deleted tab if it's been opened — otherwise its
      // first activation will load it.
      deleted.initiated
        ? loadTab("deleted", { append: false, offset: 0 })
        : Promise.resolve(),
    ]);
  }, [loadSideRequests, loadTab, deleted.initiated]);

  const handleDelete = async (download: { id: number }) => {
    setWorkingId(download.id);
    setMessage("");
    setErrorMessage("");
    try {
      const response = await axios.post(
        `${domain}/api/downloads/${download.id}/delete`,
        {},
        { withCredentials: true }
      );
      const action: string | undefined = response.data?.status;
      if (action === "hit_and_run") {
        setMessage(
          "Marked as Hit & Run — torrent will auto-delete once seeding requirement is met"
        );
      } else if (response.status === 202) {
        setMessage("Delete request submitted");
      } else {
        setMessage("Torrent deleted");
      }
      await reloadAll();
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
      await reloadAll();
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

  const handleSortFieldChange = (value: DownloadSortField) => {
    setSortField(value);
    if (value === "createdAt" || value === "torrentSize" || value === "deletedAt") {
      setSortDirection("desc");
      return;
    }
    setSortDirection("asc");
  };

  const userOptions = useMemo(() => {
    return availableUsers
      .filter((u) => u && u.trim())
      .map((u) => ({ value: u, label: u }))
      .sort((a, b) =>
        a.label.localeCompare(b.label, undefined, { sensitivity: "base" })
      );
  }, [availableUsers]);

  const canFilterByUser = isAdmin && userOptions.length > 1;
  const canSortByUser = isAdmin && userOptions.length > 1;

  const tabState = activeTab === "installed" ? installed : deleted;
  const hasMore = tabState.rows.length < tabState.total;

  const renderDownloadRow = (download: DownloadRecord) => {
    const progress = Math.max(0, Math.min(100, download.progressPercent || 0));
    const safeIn = download.safeToDeleteAt
      ? formatCountdown(download.safeToDeleteAt)
      : null;
    const deleteLabel = isAdmin ? "Delete Torrent" : "Request Delete";
    const actionElement = download.deletedAt ? (
      <span className="shrink-0 text-xs opacity-70">
        Deleted by {download.deletedByUsername || "unknown"} at{" "}
        {formatDate(download.deletedAt)}
      </span>
    ) : (
      <button
        className="btn btn-error btn-sm shrink-0"
        disabled={workingId === download.id}
        onClick={() => handleDelete(download)}
      >
        {deleteLabel}
      </button>
    );

    return (
      <div
        key={download.id}
        className="flex flex-col gap-3 rounded-xl border border-base-300 bg-base-200 p-3 sm:flex-row sm:items-center"
      >
        <div className="min-w-0 flex-1">
          <div className="mb-1 flex flex-wrap items-start justify-between gap-2">
            <p className="min-w-0 font-semibold break-all">
              {download.filename || download.fid}
              {download.isFreeleech ? (
                <span className="badge badge-warning badge-sm ml-2 align-middle">
                  FREELEECH
                </span>
              ) : null}
              {safeIn && safeIn !== "now" && !download.deletedAt ? (
                <span className="badge badge-outline badge-sm ml-2 align-middle">
                  Safe in {safeIn}
                </span>
              ) : null}
            </p>
            <div className="flex gap-2">
              {download.hasHitAndRun && !download.deletedAt ? (
                <span className="badge badge-warning badge-sm">Hit &amp; Run</span>
              ) : null}
              {download.hasPendingDeleteRequest && !download.deletedAt ? (
                <span className="badge badge-info badge-sm">Delete Pending</span>
              ) : null}
              {download.deletedAt ? (
                <span className="badge badge-neutral badge-sm">Deleted</span>
              ) : null}
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            {!download.deletedAt ? (
              <div className="flex shrink-0 items-center gap-2">
                <span className="text-xs tabular-nums opacity-75">
                  {progress.toFixed(0)}%
                </span>
                <progress
                  className="progress progress-info h-2 w-32"
                  value={progress}
                  max={100}
                />
              </div>
            ) : null}
            <p className="w-full min-w-0 text-xs opacity-70 sm:w-auto">
              Added: {formatDate(download.createdAt)} - Size:{" "}
              {formatBytes(download.torrentSize || 0)}
              {!download.deletedAt
                ? ` - State: ${normalizeState(download.qbtState)}`
                : ""}
              {canSortByUser ? ` - User: ${download.username}` : ""}
            </p>
            <div className="ml-auto sm:hidden">{actionElement}</div>
          </div>
        </div>

        <div className="hidden sm:block">{actionElement}</div>
      </div>
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Torrents</h1>
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

      {sideLoading ? (
        <div className="skeleton h-24 w-full"></div>
      ) : (
        <>
          {hitAndRunRequests.length > 0 ? (
            <div className="space-y-3">
              <h2 className="text-xl font-semibold">
                Hit &amp; Run{isAdmin ? "" : " — your torrents"}
              </h2>
              <p className="text-xs opacity-70">
                Torrents queued for deletion that have not finished seeding. They
                auto-delete once the seeding window passes (168h after completion,
                +24h grace).
              </p>
              {hitAndRunRequests.map((request) => {
                const safeIn = formatCountdown(request.safeToDeleteAt);
                const autoIn = formatCountdown(request.autoDeleteAt);
                return (
                  <div
                    key={`hnr-${request.id}`}
                    className="rounded-xl border border-warning bg-base-200 p-4"
                  >
                    <div className="flex flex-wrap items-start justify-between gap-2">
                      <div className="min-w-0">
                        <p className="font-semibold break-all">
                          {request.downloadFilename ||
                            request.downloadFid ||
                            `Download #${request.downloadEventID}`}
                        </p>
                        <p className="text-xs opacity-70">
                          Requested by {request.requestedByUsername} at{" "}
                          {formatDate(request.createdAt)}
                          {request.downloadSize
                            ? ` - Size: ${formatBytes(request.downloadSize)}`
                            : ""}
                        </p>
                        <p className="text-xs opacity-80">
                          Safe to delete in:{" "}
                          <span className="font-mono">{safeIn}</span>
                          {" · "}
                          Auto-deletes in:{" "}
                          <span className="font-mono">{autoIn}</span>
                        </p>
                        {request.reason ? (
                          <p className="mt-1 text-sm">Reason: {request.reason}</p>
                        ) : null}
                      </div>
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="badge badge-warning badge-sm">
                          Hit &amp; Run
                        </span>
                        {request.downloadIsFreeleech ? (
                          <span className="badge badge-warning badge-sm">
                            FREELEECH
                          </span>
                        ) : null}
                      </div>
                    </div>
                    {isAdmin ? (
                      <button
                        className="btn btn-error btn-sm mt-3"
                        disabled={workingId === request.downloadEventID}
                        onClick={() =>
                          handleDelete({ id: request.downloadEventID })
                        }
                      >
                        Force Delete Now
                      </button>
                    ) : null}
                  </div>
                );
              })}
            </div>
          ) : null}

          {isAdmin && pendingRequests.length > 0 ? (
            <div className="space-y-3">
              <h2 className="text-xl font-semibold">Pending Delete Requests</h2>
              {pendingRequests.map((request) => (
                <div
                  key={`pending-${request.id}`}
                  className="rounded-xl border border-base-300 bg-base-200 p-4"
                >
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div className="min-w-0">
                      <p className="font-semibold break-all">
                        {request.downloadFilename ||
                          request.downloadFid ||
                          `Download #${request.downloadEventID}`}
                      </p>
                      <p className="text-xs opacity-70">
                        Request #{request.id} for Download #
                        {request.downloadEventID}
                        {request.downloadSize
                          ? ` - Size: ${formatBytes(request.downloadSize)}`
                          : ""}
                      </p>
                      <p className="text-xs opacity-70">
                        Requested by {request.requestedByUsername} at{" "}
                        {formatDate(request.createdAt)}
                      </p>
                    </div>
                    {request.downloadIsFreeleech ? (
                      <span className="badge badge-warning badge-sm">FREELEECH</span>
                    ) : null}
                  </div>
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
              ))}
            </div>
          ) : null}
        </>
      )}

      <div role="tablist" className="tabs tabs-bordered">
        <button
          role="tab"
          className={`tab ${activeTab === "installed" ? "tab-active" : ""}`}
          onClick={() => setActiveTab("installed")}
        >
          Installed{installed.initiated ? ` (${installed.total})` : ""}
        </button>
        <button
          role="tab"
          className={`tab ${activeTab === "deleted" ? "tab-active" : ""}`}
          onClick={() => setActiveTab("deleted")}
        >
          Delete History{deleted.initiated ? ` (${deleted.total})` : ""}
        </button>
      </div>

      <div className="rounded-xl border border-base-300 bg-base-200 p-4">
        <div className="flex flex-wrap items-end gap-3">
          <label className="form-control w-full sm:max-w-md">
            <span className="label-text text-xs uppercase tracking-wide opacity-70">
              Search
            </span>
            <input
              type="text"
              className="input input-bordered input-sm"
              placeholder="Search title or fid"
              value={searchTerm}
              onChange={(event) => setSearchTerm(event.target.value)}
            />
          </label>

          {canFilterByUser ? (
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
                {userOptions.map((user) => (
                  <option key={user.value} value={user.value}>
                    {user.label}
                  </option>
                ))}
              </select>
            </label>
          ) : null}

          <label className="form-control w-full sm:w-56">
            <span className="label-text text-xs uppercase tracking-wide opacity-70">
              Sort by
            </span>
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
              {activeTab === "deleted" ? (
                <option value="deletedAt">Deleted date</option>
              ) : null}
              {canSortByUser ? <option value="username">User</option> : null}
            </select>
          </label>

          <button
            className="btn btn-outline btn-sm"
            onClick={() =>
              setSortDirection(sortDirection === "asc" ? "desc" : "asc")
            }
          >
            {sortDirection === "asc" ? "Ascending" : "Descending"}
          </button>
        </div>

        <p className="mt-2 text-xs opacity-70">
          Showing {tabState.rows.length} of {tabState.total} torrents
        </p>
      </div>

      {tabState.loading && tabState.rows.length === 0 ? (
        <div className="space-y-3">
          {[...Array(6).keys()].map((index) => (
            <div key={index} className="skeleton h-24 w-full"></div>
          ))}
        </div>
      ) : tabState.rows.length === 0 ? (
        <div className="rounded-xl border border-base-300 bg-base-200 p-5 text-sm opacity-80">
          {activeTab === "installed"
            ? "No installed torrents match your filters."
            : "No deleted torrents match your filters."}
        </div>
      ) : (
        <div className="space-y-3">
          {tabState.rows.map(renderDownloadRow)}
          {hasMore ? (
            <div className="flex justify-center pt-2">
              <button
                className="btn btn-outline btn-sm"
                disabled={tabState.loading}
                onClick={() =>
                  loadTab(activeTab, {
                    append: true,
                    offset: tabState.offset,
                  })
                }
              >
                {tabState.loading
                  ? "Loading..."
                  : `Load more (${tabState.total - tabState.rows.length} remaining)`}
              </button>
            </div>
          ) : null}
        </div>
      )}
    </div>
  );
}
