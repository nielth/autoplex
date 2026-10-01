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
  uploaded: number;
  upSpeed: number;
  ratio: number;
  seeds: number;
  seedsInSwarm: number;
  savePath?: string;
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

const PAGE_SIZE = 100;

type ColumnKey =
  | "user"
  | "added"
  | "size"
  | "progress"
  | "state"
  | "completed"
  | "uploaded"
  | "upSpeed"
  | "ratio"
  | "seeds"
  | "savePath"
  | "deleted"
  | "deletedBy";

interface ColumnDef {
  key: ColumnKey;
  label: string;
  // Fixed width in px, so columns never shift when sorting or data changes.
  width: number;
  sort?: DownloadSortField;
  adminOnly?: boolean;
  tab?: TabKey;
}

// Name and the delete button are always shown; these can be toggled.
const COLUMNS: ColumnDef[] = [
  { key: "user", label: "User", width: 110, sort: "username", adminOnly: true },
  { key: "added", label: "Added", width: 165, sort: "createdAt" },
  { key: "size", label: "Size", width: 100, sort: "torrentSize" },
  { key: "progress", label: "Progress", width: 140, tab: "installed" },
  { key: "state", label: "State", width: 120, tab: "installed" },
  { key: "uploaded", label: "Uploaded", width: 100, tab: "installed" },
  { key: "upSpeed", label: "Up speed", width: 100, tab: "installed" },
  { key: "ratio", label: "Ratio", width: 70, tab: "installed" },
  { key: "seeds", label: "Seeds", width: 90, tab: "installed" },
  { key: "completed", label: "Completed", width: 165, tab: "installed" },
  { key: "savePath", label: "Save path", width: 220, tab: "installed" },
  { key: "deleted", label: "Deleted", width: 165, sort: "deletedAt", tab: "deleted" },
  { key: "deletedBy", label: "Deleted by", width: 120, tab: "deleted" },
];

const NAME_MIN_WIDTH = 360;
const ACTION_WIDTH = 135;

const HIDDEN_COLUMNS_KEY = "downloads.hiddenColumns";

function loadHiddenColumns(): Partial<Record<ColumnKey, boolean>> {
  const fallback = { completed: true, ratio: true };
  try {
    const stored = localStorage.getItem(HIDDEN_COLUMNS_KEY);
    return stored ? JSON.parse(stored) : fallback;
  } catch {
    return fallback;
  }
}

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

// qBittorrent's API states, labelled the way qBittorrent's own web UI shows
// them. "missing" and "deleted" come from autoplex itself.
const STATE_LABELS: Record<string, string> = {
  downloading: "Downloading",
  forcedDL: "[F] Downloading",
  metaDL: "Downloading metadata",
  forcedMetaDL: "[F] Downloading metadata",
  stalledDL: "Stalled",
  queuedDL: "Queued",
  pausedDL: "Paused",
  stoppedDL: "Stopped",
  uploading: "Seeding",
  stalledUP: "Seeding",
  forcedUP: "[F] Seeding",
  queuedUP: "Queued",
  pausedUP: "Completed",
  stoppedUP: "Completed",
  checkingDL: "Checking",
  checkingUP: "Checking",
  checkingResumeData: "Checking resume data",
  allocating: "Allocating",
  moving: "Moving",
  missingFiles: "Missing files",
  error: "Errored",
  missing: "Not in qBittorrent",
  deleted: "Deleted",
};

function normalizeState(state?: string) {
  if (!state) return "Unknown";
  return STATE_LABELS[state] ?? state;
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
  const [isImporting, setIsImporting] = useState<boolean>(false);
  const [hiddenColumns, setHiddenColumns] =
    useState<Partial<Record<ColumnKey, boolean>>>(loadHiddenColumns);
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
      if (response.data?.status === "hit_and_run_queued") {
        setMessage(
          "Still seeding — queued for deletion, it auto-deletes once the seeding window passes"
        );
      } else if (response.status === 202) {
        setMessage("Delete request submitted — torrent paused if it was still downloading");
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

  const handleImport = async () => {
    setIsImporting(true);
    setMessage("");
    setErrorMessage("");
    try {
      const response = await axios.post(
        `${domain}/api/downloads/import-qbt`,
        {},
        { withCredentials: true }
      );
      const imported: number = response.data?.imported ?? 0;
      setMessage(
        imported > 0
          ? `Imported ${imported} torrent(s) from qBittorrent`
          : "Everything in qBittorrent is already tracked"
      );
      await reloadAll();
    } catch (error) {
      const response = axios.isAxiosError(error) ? error.response : undefined;
      if (response?.status === 401) {
        await authProvider.signout();
        navigate("/login");
        return;
      }
      setErrorMessage(response?.data?.error || "Import from qBittorrent failed");
    } finally {
      setIsImporting(false);
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

  const sentinelRef = useRef<HTMLDivElement | null>(null);
  useEffect(() => {
    if (!hasMore) return;
    const node = sentinelRef.current;
    if (!node) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (!entry?.isIntersecting) return;
        if (tabState.loading) return;
        loadTab(activeTab, { append: true, offset: tabState.offset });
      },
      { rootMargin: "400px" }
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [hasMore, tabState.loading, tabState.offset, activeTab, loadTab]);

  const visibleColumns = COLUMNS.filter(
    (column) =>
      hiddenColumns[column.key] !== true &&
      (!column.adminOnly || canSortByUser) &&
      (!column.tab || column.tab === activeTab)
  );

  const toggleColumn = (key: ColumnKey) => {
    setHiddenColumns((previous) => {
      const next = { ...previous, [key]: !previous[key] };
      try {
        localStorage.setItem(HIDDEN_COLUMNS_KEY, JSON.stringify(next));
      } catch {
        // Column choice is a convenience, the page works without storage.
      }
      return next;
    });
  };

  const handleHeaderSort = (field: DownloadSortField) => {
    if (field === sortField) {
      setSortDirection(sortDirection === "asc" ? "desc" : "asc");
      return;
    }
    handleSortFieldChange(field);
  };

  const renderSortHeader = (label: string, field?: DownloadSortField) => {
    if (!field) return label;
    const isSorted = sortField === field;
    return (
      <button
        type="button"
        className="font-semibold hover:underline"
        onClick={() => handleHeaderSort(field)}
      >
        {label}
        <span className={`ml-1 inline-block w-3 ${isSorted ? "" : "invisible"}`}>
          {isSorted && sortDirection === "asc" ? "▲" : "▼"}
        </span>
      </button>
    );
  };

  const renderCell = (column: ColumnKey, download: DownloadRecord) => {
    const progress = Math.max(0, Math.min(100, download.progressPercent || 0));
    switch (column) {
      case "user":
        return download.username;
      case "added":
        return formatDate(download.createdAt);
      case "size":
        return formatBytes(download.torrentSize || 0);
      case "progress":
        return (
          <div className="flex items-center gap-2">
            <progress className="progress progress-info h-2 w-20" value={progress} max={100} />
            <span className="tabular-nums opacity-75">{progress.toFixed(0)}%</span>
          </div>
        );
      case "state":
        return normalizeState(download.qbtState);
      case "completed":
        return formatDate(download.completedAt);
      case "uploaded":
        return formatBytes(download.uploaded || 0);
      case "upSpeed":
        return download.upSpeed > 0 ? `${formatBytes(download.upSpeed)}/s` : "-";
      case "ratio":
        return download.ratio >= 0 ? download.ratio.toFixed(2) : "-";
      case "seeds":
        return `${download.seeds ?? 0} (${download.seedsInSwarm ?? 0})`;
      case "savePath":
        return download.savePath || "-";
      case "deleted":
        return formatDate(download.deletedAt);
      case "deletedBy":
        return download.deletedByUsername || "-";
    }
  };

  // Shown while the tracker's seeding window is still running. With showSafe
  // a "SAFE" badge is shown once it has passed (used in the compact list).
  const renderSeedingBadge = (download: DownloadRecord, showSafe = false) => {
    if (download.deletedAt || !download.safeToDeleteAt) return null;
    const safeIn = formatCountdown(download.safeToDeleteAt);
    if (safeIn === "-") return null;
    if (safeIn === "now") {
      return showSafe ? (
        <span className="badge badge-success badge-outline badge-xs ml-2">SAFE</span>
      ) : null;
    }
    return (
      <span
        className="badge badge-info badge-outline badge-xs ml-2"
        title={`Must keep seeding: safe to delete in ${safeIn}`}
      >
        SEEDING {safeIn}
      </span>
    );
  };

  const renderAction = (download: DownloadRecord) => {
    if (download.deletedAt) return null;
    const safeIn = download.safeToDeleteAt ? formatCountdown(download.safeToDeleteAt) : null;
    const isStillSeeding = Boolean(safeIn) && safeIn !== "now";
    const deleteLabel = isAdmin ? (isStillSeeding ? "Queue delete" : "Delete") : "Request delete";
    return (
      <button
        className="btn btn-error btn-xs whitespace-nowrap"
        disabled={workingId === download.id}
        onClick={() => handleDelete(download)}
      >
        {deleteLabel}
      </button>
    );
  };

  const renderRequestTable = (
    title: string,
    description: string,
    requests: DeleteRequestRecord[],
    actionLabel: string,
    actionClass: string
  ) => (
    <div className="space-y-2">
      <h2 className="text-lg font-semibold">{title}</h2>
      <p className="text-xs opacity-70">{description}</p>
      <div className="overflow-x-auto rounded-xl border border-base-300">
        <table className="table table-xs">
          <thead>
            <tr>
              <th>Name</th>
              <th>Requested by</th>
              <th>Requested</th>
              <th>Size</th>
              <th>Safe to delete in</th>
              <th>Auto-deletes in</th>
              <th>Reason</th>
              {isAdmin ? <th></th> : null}
            </tr>
          </thead>
          <tbody>
            {requests.map((request) => (
              <tr key={`request-${request.id}`} className="hover">
                <td className="min-w-64 break-all">
                  {request.downloadFilename ||
                    request.downloadFid ||
                    `Download #${request.downloadEventID}`}
                  {request.downloadIsFreeleech ? (
                    <span className="badge badge-warning badge-xs ml-2">FREELEECH</span>
                  ) : null}
                </td>
                <td>{request.requestedByUsername}</td>
                <td className="whitespace-nowrap">{formatDate(request.createdAt)}</td>
                <td className="whitespace-nowrap">
                  {request.downloadSize ? formatBytes(request.downloadSize) : "-"}
                </td>
                <td className="whitespace-nowrap font-mono">
                  {formatCountdown(request.safeToDeleteAt)}
                </td>
                <td className="whitespace-nowrap font-mono">
                  {formatCountdown(request.autoDeleteAt)}
                </td>
                <td>{request.reason || "-"}</td>
                {isAdmin ? (
                  <td>
                    <button
                      className={`btn btn-xs ${actionClass}`}
                      disabled={workingId === request.id}
                      onClick={() => handleApprove(request.id)}
                    >
                      {actionLabel}
                    </button>
                  </td>
                ) : null}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Torrents</h1>
          <p className="text-sm opacity-70">
            {isAdmin
              ? "All tracked torrents across users"
              : "Your tracked torrents and delete options"}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          {isAdmin ? (
            <button
              className="btn btn-outline btn-sm"
              disabled={isImporting}
              onClick={handleImport}
              title="Add torrents that are in qBittorrent but not tracked here, under your user"
            >
              {isImporting ? "Importing..." : "Import from qBittorrent"}
            </button>
          ) : null}
          <button
            className="btn btn-primary btn-sm"
            disabled={isScanningPlex}
            onClick={handlePlexScan}
          >
            {isScanningPlex ? "Starting Scan..." : "Scan Plex: Movies + TV Shows"}
          </button>
        </div>
      </div>

      {message ? <div className="alert alert-success">{message}</div> : null}
      {errorMessage ? <div className="alert alert-error">{errorMessage}</div> : null}

      {sideLoading ? (
        <div className="skeleton h-16 w-full"></div>
      ) : (
        <>
          {hitAndRunRequests.length > 0
            ? renderRequestTable(
                `Hit & Run${isAdmin ? "" : " — your torrents"}`,
                "Queued for deletion but not done seeding. They auto-delete once the seeding window passes (168h after completion, +24h grace).",
                hitAndRunRequests,
                "Force delete now",
                "btn-error"
              )
            : null}
          {isAdmin && pendingRequests.length > 0
            ? renderRequestTable(
                "Pending delete requests",
                "Delete requests from users, waiting for approval.",
                pendingRequests,
                "Approve and delete",
                "btn-success"
              )
            : null}
        </>
      )}

      <div className="flex flex-wrap items-center gap-3">
        <div role="tablist" className="tabs tabs-boxed tabs-sm">
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
            Delete history{deleted.initiated ? ` (${deleted.total})` : ""}
          </button>
        </div>

        <input
          type="text"
          className="input input-bordered input-sm w-full sm:w-80"
          placeholder="Search title or fid"
          value={searchTerm}
          onChange={(event) => setSearchTerm(event.target.value)}
        />

        {canFilterByUser ? (
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
        ) : null}

        <div className="dropdown dropdown-end ml-auto hidden lg:block">
          <div tabIndex={0} role="button" className="btn btn-outline btn-sm">
            Columns
          </div>
          <ul
            tabIndex={0}
            className="menu dropdown-content z-10 w-52 rounded-box bg-base-200 p-2 shadow"
          >
            {COLUMNS.filter(
              (column) =>
                (!column.adminOnly || canSortByUser) &&
                (!column.tab || column.tab === activeTab)
            ).map((column) => (
              <li key={column.key}>
                <label className="label cursor-pointer justify-start gap-3">
                  <input
                    type="checkbox"
                    className="checkbox checkbox-sm"
                    checked={hiddenColumns[column.key] !== true}
                    onChange={() => toggleColumn(column.key)}
                  />
                  <span className="label-text">{column.label}</span>
                </label>
              </li>
            ))}
          </ul>
        </div>

        <span className="text-xs opacity-70">
          {tabState.rows.length} of {tabState.total}
        </span>
      </div>

      {tabState.loading && tabState.rows.length === 0 ? (
        <div className="skeleton h-64 w-full"></div>
      ) : tabState.rows.length === 0 ? (
        <div className="rounded-xl border border-base-300 bg-base-200 p-5 text-sm opacity-80">
          {activeTab === "installed"
            ? "No installed torrents match your filters."
            : "No deleted torrents match your filters."}
        </div>
      ) : (
        <>
        {/* Compact list for phones and small screens */}
        <div className="space-y-2 lg:hidden">
          {tabState.rows.map((download) => {
            const progress = Math.max(0, Math.min(100, download.progressPercent || 0));
            return (
              <div key={download.id} className="rounded-lg border border-base-300 bg-base-200 p-3">
                <p className="break-all text-sm font-medium">
                  {download.filename || download.fid}
                  {download.isFreeleech ? (
                    <span className="badge badge-warning badge-xs ml-2">FREELEECH</span>
                  ) : null}
                  {renderSeedingBadge(download, true)}
                  {!download.deletedAt && download.hasHitAndRun ? (
                    <span className="badge badge-warning badge-xs ml-2">Hit &amp; Run</span>
                  ) : null}
                  {!download.deletedAt && download.hasPendingDeleteRequest ? (
                    <span className="badge badge-info badge-xs ml-2">Delete pending</span>
                  ) : null}
                </p>
                <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs">
                  <span className="opacity-70">{download.username}</span>
                  <span className="tabular-nums opacity-70">
                    {formatBytes(download.torrentSize || 0)}
                  </span>
                  {download.deletedAt ? (
                    <span className="opacity-70">Deleted {formatDate(download.deletedAt)}</span>
                  ) : (
                    <span className="flex items-center gap-1">
                      <progress className="progress progress-info h-2 w-16" value={progress} max={100} />
                      <span className="tabular-nums opacity-70">{progress.toFixed(0)}%</span>
                    </span>
                  )}
                  <span className="ml-auto">{renderAction(download)}</span>
                </div>
              </div>
            );
          })}
        </div>

        <div className="hidden overflow-x-auto rounded-xl border border-base-300 lg:block">
          <table
            className="table table-sm table-pin-rows table-fixed"
            style={{
              minWidth:
                NAME_MIN_WIDTH +
                visibleColumns.reduce((sum, column) => sum + column.width, 0) +
                (activeTab === "installed" ? ACTION_WIDTH : 0),
            }}
          >
            <colgroup>
              <col />
              {visibleColumns.map((column) => (
                <col key={column.key} style={{ width: column.width }} />
              ))}
              {activeTab === "installed" ? <col style={{ width: ACTION_WIDTH }} /> : null}
            </colgroup>
            <thead>
              <tr className="bg-base-200">
                <th>{renderSortHeader("Name", "filename")}</th>
                {visibleColumns.map((column) => (
                  <th key={column.key} className="truncate">
                    {renderSortHeader(column.label, column.sort)}
                  </th>
                ))}
                {activeTab === "installed" ? <th></th> : null}
              </tr>
            </thead>
            <tbody>
              {tabState.rows.map((download) => (
                <tr key={download.id} className="hover">
                  <td className="break-all">
                    {download.filename || download.fid}
                    {download.isFreeleech ? (
                      <span className="badge badge-warning badge-xs ml-2">FREELEECH</span>
                    ) : null}
                    {renderSeedingBadge(download)}
                    {!download.deletedAt && download.hasHitAndRun ? (
                      <span className="badge badge-warning badge-xs ml-2">Hit &amp; Run</span>
                    ) : null}
                    {!download.deletedAt && download.hasPendingDeleteRequest ? (
                      <span className="badge badge-info badge-xs ml-2">Delete pending</span>
                    ) : null}
                  </td>
                  {visibleColumns.map((column) => (
                    <td
                      key={column.key}
                      className="truncate tabular-nums"
                      title={column.key === "savePath" ? download.savePath : undefined}
                    >
                      {renderCell(column.key, download)}
                    </td>
                  ))}
                  {activeTab === "installed" ? <td>{renderAction(download)}</td> : null}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {hasMore ? (
          <div ref={sentinelRef} className="flex justify-center py-4 text-xs opacity-60">
            {tabState.loading
              ? "Loading more..."
              : `${tabState.total - tabState.rows.length} more`}
          </div>
        ) : null}
        </>
      )}
    </div>
  );
}
