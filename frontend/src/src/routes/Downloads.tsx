import axios from "axios";
import { useEffect, useState } from "react";
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
  return new Date(value).toLocaleString();
}

function normalizeState(state?: string) {
  if (!state) {
    return "unknown";
  }
  return state.replace(/_/g, " ");
}

export function Downloads() {
  const [downloads, setDownloads] = useState<DownloadRecord[]>([]);
  const [pendingRequests, setPendingRequests] = useState<DeleteRequestRecord[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [workingId, setWorkingId] = useState<number | null>(null);
  const [message, setMessage] = useState<string>("");
  const [errorMessage, setErrorMessage] = useState<string>("");
  const navigate = useNavigate();
  const domain = getApiDomain();
  const isAdmin = authProvider.isAdmin;

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
      <div>
        <h1 className="text-2xl font-semibold">Installed Torrents</h1>
        <p className="text-sm opacity-70">
          {isAdmin
            ? "All tracked torrents across users"
            : "Your tracked torrents and delete options"}
        </p>
      </div>

      {message ? <div className="alert alert-success">{message}</div> : null}
      {errorMessage ? <div className="alert alert-error">{errorMessage}</div> : null}

      {downloads.length === 0 ? (
        <div className="rounded-xl border border-base-300 bg-base-200 p-5 text-sm opacity-80">
          No tracked torrents found.
        </div>
      ) : (
        <div className="space-y-3">
          {downloads.map((download) => {
            const progress = Math.max(0, Math.min(100, download.progressPercent || 0));
            const deleteLabel =
              isAdmin || download.isFreeleech ? "Delete Torrent" : "Request Delete";

            return (
              <div
                key={download.id}
                className="rounded-xl border border-base-300 bg-base-200 p-4"
              >
                <div className="mb-2 flex flex-wrap items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="font-semibold break-all">{download.filename || download.fid}</p>
                    <p className="text-xs opacity-70">
                      Added: {formatDate(download.createdAt)}
                      {isAdmin ? ` - User: ${download.username}` : ""}
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

                <div className="mb-2 text-xs opacity-75">
                  Size: {formatBytes(download.torrentSize || 0)} - State:{" "}
                  {normalizeState(download.qbtState)}
                </div>

                <progress
                  className="progress progress-info w-full"
                  value={progress}
                  max={100}
                />
                <p className="mt-1 text-xs opacity-75">{progress.toFixed(2)}%</p>

                {download.deletedAt ? (
                  <p className="mt-2 text-xs opacity-70">
                    Deleted by {download.deletedByUsername || "unknown"} at{" "}
                    {formatDate(download.deletedAt)}
                  </p>
                ) : (
                  <div className="mt-3">
                    <button
                      className="btn btn-error btn-sm"
                      disabled={workingId === download.id}
                      onClick={() => handleDelete(download)}
                    >
                      {deleteLabel}
                    </button>
                  </div>
                )}
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
