import axios from "axios";
import { useCallback, useEffect, useState } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { authProvider } from "../auth";
import { formatBytes } from "../scripts/formatBytes";
import { getApiDomain } from "../scripts/getApiDomain";

type BalancerDisk = {
  name: string;
  role: "spare" | "pool";
  free: number;
  total: number;
  heavyCount: number;
  recentHeavyCount: number;
};

type PlanItem = {
  name: string;
  from: string;
  to: string;
  size: number;
  mbps: number;
  reason: string;
  pending: boolean;
};

type BitrateRow = {
  name: string;
  disk: string;
  size: number;
  mbps: number;
  heavy: boolean;
};

type BalancerStatus = {
  running: boolean;
  checkedAt: string;
  message: string;
  heavyMbps: number;
  disks: BalancerDisk[] | null;
  plan: PlanItem[] | null;
  topBitrates: BitrateRow[] | null;
};

type MoveRecord = {
  id: number;
  torrentName: string;
  fromDisk: string;
  toDisk: string;
  size: number;
  mbps: number;
  reason: string;
  success: boolean;
  errorMessage: string;
  createdAt: string;
};

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime()) || date.getFullYear() < 2000) return "never";
  return date.toLocaleString();
}

function formatMbps(mbps: number): string {
  return mbps > 0 ? `${mbps.toFixed(1)} Mbps` : "-";
}

// One torrent as a compact card, used instead of the tables on small screens.
function CompactRow({
  name,
  badges,
  details,
}: {
  name: string;
  badges?: React.ReactNode;
  details: React.ReactNode[];
}) {
  return (
    <div className="rounded-lg border border-base-300 bg-base-200 p-3">
      <p className="break-all text-sm font-medium">
        {name}
        {badges}
      </p>
      <div className="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-xs opacity-70">
        {details.map((detail, index) => (
          <span key={index}>{detail}</span>
        ))}
      </div>
    </div>
  );
}

export function DiskBalancer() {
  const navigate = useNavigate();
  const domain = getApiDomain();
  const [status, setStatus] = useState<BalancerStatus | null>(null);
  const [moves, setMoves] = useState<MoveRecord[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [errorMessage, setErrorMessage] = useState<string>("");

  const load = useCallback(async () => {
    try {
      const resp = await axios.get(`${domain}/api/disk-balancer`, {
        withCredentials: true,
      });
      setStatus(resp.data.status);
      setMoves(resp.data.moves ?? []);
      setErrorMessage("");
    } catch (error) {
      const response = axios.isAxiosError(error) ? error.response : undefined;
      if (response?.status === 401) {
        await authProvider.signout();
        navigate("/login");
        return;
      }
      setErrorMessage(response?.data?.error || "Failed to load disk balancer");
    } finally {
      setLoading(false);
    }
  }, [domain, navigate]);

  useEffect(() => {
    if (!authProvider.isAdmin) return;
    load();
    const id = window.setInterval(load, 60_000);
    return () => window.clearInterval(id);
  }, [load]);

  if (!authProvider.isAdmin) {
    return <Navigate to="/" replace />;
  }

  const plan = status?.plan ?? [];
  const disks = status?.disks ?? [];
  const topBitrates = status?.topBitrates ?? [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Disk Balancer</h1>
        <p className="text-sm opacity-70">
          Moves high bitrate torrents off the download disk and spreads them
          across the other disks. Checks every 10 minutes.
        </p>
      </div>

      {errorMessage ? <div className="alert alert-error">{errorMessage}</div> : null}

      {loading ? (
        <div className="skeleton h-24 w-full"></div>
      ) : status ? (
        <>
          <div className="rounded-xl bg-base-200 p-4">
            <div className="flex flex-wrap items-center gap-2">
              {status.running ? (
                <span className="badge badge-success">Active</span>
              ) : (
                <span className="badge badge-ghost">Not running</span>
              )}
              <span className="text-sm">{status.message}</span>
            </div>
            <p className="mt-1 text-xs opacity-70">
              Last checked: {formatDate(status.checkedAt)}
              {status.heavyMbps ? ` · Heavy at ${status.heavyMbps} Mbps or more` : ""}
            </p>
          </div>

          {disks.length > 0 ? (
            <div className="space-y-3">
              <h2 className="text-xl font-semibold">Disks</h2>
              <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
                {disks.map((disk) => {
                  const usedPercent =
                    disk.total > 0 ? ((disk.total - disk.free) / disk.total) * 100 : 0;
                  return (
                    <div key={disk.name} className="rounded-xl bg-base-200 p-3 sm:p-4">
                      <div className="flex items-center justify-between">
                        <span className="font-mono font-semibold">{disk.name}</span>
                        <span className="badge badge-sm">
                          {disk.role === "spare" ? "downloads" : "pool"}
                        </span>
                      </div>
                      <progress
                        className="progress progress-primary mt-2 w-full"
                        value={usedPercent}
                        max={100}
                      ></progress>
                      <p className="text-xs opacity-70">
                        {formatBytes(disk.free)} free of {formatBytes(disk.total)}
                      </p>
                      <p className="text-xs opacity-70">
                        {disk.heavyCount} heavy ({disk.recentHeavyCount} recent)
                      </p>
                    </div>
                  );
                })}
              </div>
            </div>
          ) : null}

          <div className="space-y-3">
            <h2 className="text-xl font-semibold">Plan</h2>
            {plan.length === 0 ? (
              <p className="text-sm opacity-70">Nothing planned.</p>
            ) : (
              <>
              <div className="space-y-2 lg:hidden">
                {plan.map((item, index) => (
                  <CompactRow
                    key={`${item.name}-${index}`}
                    name={item.name}
                    badges={
                      item.pending ? (
                        <span className="badge badge-ghost badge-xs ml-2">next</span>
                      ) : null
                    }
                    details={[
                      <span className="font-mono">
                        {item.from} → {item.to}
                      </span>,
                      formatBytes(item.size),
                      formatMbps(item.mbps),
                      <span className="break-all">{item.reason}</span>,
                    ]}
                  />
                ))}
              </div>
              <div className="hidden overflow-x-auto lg:block">
                <table className="table table-sm">
                  <thead>
                    <tr>
                      <th>Torrent</th>
                      <th>Move</th>
                      <th>Size</th>
                      <th>Bitrate</th>
                      <th>Reason</th>
                    </tr>
                  </thead>
                  <tbody>
                    {plan.map((item, index) => (
                      <tr key={`${item.name}-${index}`}>
                        <td className="min-w-64 break-all">{item.name}</td>
                        <td className="whitespace-nowrap font-mono">
                          {item.from} → {item.to}
                          {item.pending ? (
                            <span
                              className="badge badge-ghost badge-sm ml-2"
                              title="Moves on a later run, once room has been made"
                            >
                              next
                            </span>
                          ) : null}
                        </td>
                        <td className="whitespace-nowrap">{formatBytes(item.size)}</td>
                        <td className="whitespace-nowrap">{formatMbps(item.mbps)}</td>
                        <td className="break-all text-xs">{item.reason}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              </>
            )}
          </div>

          {topBitrates.length > 0 ? (
            <div className="space-y-3">
              <h2 className="text-xl font-semibold">Highest bitrates</h2>
              <p className="text-xs opacity-70">
                Measured with ffprobe on the biggest video file. Torrents too small
                to be heavy are guessed from their size.
              </p>
              <div className="space-y-2 lg:hidden">
                {topBitrates.map((row, index) => (
                  <CompactRow
                    key={`${row.name}-${index}`}
                    name={row.name}
                    badges={
                      row.heavy ? (
                        <span className="badge badge-warning badge-xs ml-2">heavy</span>
                      ) : null
                    }
                    details={[
                      <span className="font-mono">{row.disk}</span>,
                      formatBytes(row.size),
                      formatMbps(row.mbps),
                    ]}
                  />
                ))}
              </div>
              <div className="hidden overflow-x-auto lg:block">
                <table className="table table-sm">
                  <thead>
                    <tr>
                      <th>Torrent</th>
                      <th>Disk</th>
                      <th>Size</th>
                      <th>Bitrate</th>
                    </tr>
                  </thead>
                  <tbody>
                    {topBitrates.map((row, index) => (
                      <tr key={`${row.name}-${index}`}>
                        <td className="min-w-64 break-all">
                          {row.name}
                          {row.heavy ? (
                            <span className="badge badge-warning badge-sm ml-2">heavy</span>
                          ) : null}
                        </td>
                        <td className="font-mono">{row.disk}</td>
                        <td className="whitespace-nowrap">{formatBytes(row.size)}</td>
                        <td className="whitespace-nowrap">{formatMbps(row.mbps)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          ) : null}
        </>
      ) : null}

      <div className="space-y-3">
        <h2 className="text-xl font-semibold">Changelog</h2>
        {moves.length === 0 ? (
          <p className="text-sm opacity-70">No moves yet.</p>
        ) : (
          <>
          <div className="space-y-2 lg:hidden">
            {moves.map((move) => (
              <CompactRow
                key={move.id}
                name={move.torrentName}
                badges={
                  !move.success ? (
                    <span className="badge badge-error badge-xs ml-2" title={move.errorMessage}>
                      failed
                    </span>
                  ) : null
                }
                details={[
                  formatDate(move.createdAt),
                  <span className="font-mono">
                    {move.fromDisk} → {move.toDisk}
                  </span>,
                  formatBytes(move.size),
                  formatMbps(move.mbps),
                  ...(move.success ? [] : [<span className="text-error">{move.errorMessage}</span>]),
                ]}
              />
            ))}
          </div>
          <div className="hidden overflow-x-auto lg:block">
            <table className="table table-sm">
              <thead>
                <tr>
                  <th>When</th>
                  <th>Torrent</th>
                  <th>Move</th>
                  <th>Size</th>
                  <th>Bitrate</th>
                  <th>Reason</th>
                </tr>
              </thead>
              <tbody>
                {moves.map((move) => (
                  <tr key={move.id}>
                    <td className="whitespace-nowrap text-xs">{formatDate(move.createdAt)}</td>
                    <td className="min-w-64 break-all">
                      {move.torrentName}
                      {!move.success ? (
                        <p className="text-xs text-error">Failed: {move.errorMessage}</p>
                      ) : null}
                    </td>
                    <td className="whitespace-nowrap font-mono">
                      {move.fromDisk} → {move.toDisk}
                    </td>
                    <td className="whitespace-nowrap">{formatBytes(move.size)}</td>
                    <td className="whitespace-nowrap">{formatMbps(move.mbps)}</td>
                    <td className="break-all text-xs">{move.reason}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          </>
        )}
      </div>
    </div>
  );
}
