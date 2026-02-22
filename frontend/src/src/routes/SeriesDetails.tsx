import axios from "axios";
import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import { authProvider } from "../auth";
import { getApiDomain } from "../scripts/getApiDomain";

interface TvMazeImage {
  medium?: string;
  original?: string;
}

interface TvMazeSchedule {
  time?: string;
  days?: string[];
}

interface TvMazeShow {
  id: number;
  name: string;
  status?: string;
  premiered?: string;
  schedule?: TvMazeSchedule;
  image?: TvMazeImage;
  summary?: string;
}

interface TvMazeSeason {
  id: number;
  number: number;
  premiereDate?: string;
  endDate?: string;
  episodeOrder?: number;
}

interface TvMazeEpisode {
  id: number;
  name: string;
  season: number;
  number: number;
  airdate?: string;
  airtime?: string;
  airstamp?: string;
}

interface TvShowSubscriptionRecord {
  id: number;
  preferredQuality: string;
  autoInstallUpcoming: boolean;
  enabled: boolean;
}

interface TvEpisodeJobRecord {
  id: number;
  tvmazeEpisodeID: number;
  status: string;
  nextCheckAt: string;
  lastError?: string;
  preferredQuality: string;
}

interface TvShowInstallStatus {
  subscription?: TvShowSubscriptionRecord;
  jobs: TvEpisodeJobRecord[];
  pendingCount: number;
  downloadedCount: number;
  failedCount: number;
}

interface SeriesDetailResponse {
  show: TvMazeShow;
  seasons: TvMazeSeason[];
  episodes: TvMazeEpisode[];
  installStatus: TvShowInstallStatus;
}

interface QueueResult {
  queued: number;
  skipped: number;
  triggered: number;
}

interface SeasonProgress {
  aired: number;
  total: number;
  complete: boolean;
}

const FALLBACK_POSTER =
  "https://static.tvmaze.com/images/no-img/no-img-portrait-text.png";

function statusBadgeClass(status?: string): string {
  const clean = (status || "").toLowerCase();
  if (clean === "running") {
    return "badge-success";
  }
  if (clean === "ended") {
    return "badge-neutral";
  }
  if (clean === "in development") {
    return "badge-warning";
  }
  return "badge-ghost";
}

function jobBadgeClass(status?: string): string {
  const clean = (status || "").toLowerCase();
  if (clean === "downloaded") {
    return "badge-success";
  }
  if (clean === "failed") {
    return "badge-error";
  }
  if (clean === "searching") {
    return "badge-info";
  }
  return "badge-ghost";
}

function isEpisodeAired(episode: TvMazeEpisode, nowMs: number, today: string): boolean {
  if (episode.airstamp) {
    const airstampMs = Date.parse(episode.airstamp);
    return !Number.isNaN(airstampMs) && airstampMs <= nowMs;
  }

  return Boolean(episode.airdate && episode.airdate <= today);
}

export function SeriesDetails() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const domain = getApiDomain();

  const [data, setData] = useState<SeriesDetailResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [quality, setQuality] = useState<string>("1080");
  const [autoInstallUpcoming, setAutoInstallUpcoming] = useState<boolean>(false);
  const [pendingAutoInstallUpcoming, setPendingAutoInstallUpcoming] = useState<
    boolean | null
  >(null);
  const [showAutoInstallModal, setShowAutoInstallModal] = useState<boolean>(false);
  const [workingAction, setWorkingAction] = useState<string>("");
  const [message, setMessage] = useState<string>("");
  const [error, setError] = useState<string>("");

  const showID = Number(id);

  const jobsByEpisodeID = useMemo(() => {
    const map = new Map<number, TvEpisodeJobRecord>();
    data?.installStatus?.jobs?.forEach((job) => {
      if (job.preferredQuality !== quality) {
        return;
      }
      const existing = map.get(job.tvmazeEpisodeID);
      if (!existing || job.id > existing.id) {
        map.set(job.tvmazeEpisodeID, job);
      }
    });
    return map;
  }, [data, quality]);

  const downloadedEpisodeIDsForQuality = useMemo(() => {
    const installed = new Set<number>();
    data?.installStatus?.jobs?.forEach((job) => {
      if (job.preferredQuality === quality && job.status === "downloaded") {
        installed.add(job.tvmazeEpisodeID);
      }
    });
    return installed;
  }, [data, quality]);

  const seasonProgressByNumber = useMemo(() => {
    const progressBySeason = new Map<number, SeasonProgress>();
    if (!data) {
      return progressBySeason;
    }

    const now = Date.now();
    const today = new Date().toISOString().slice(0, 10);
    const episodeStatsBySeason = new Map<number, { known: number; aired: number }>();

    data.episodes.forEach((episode) => {
      const stats = episodeStatsBySeason.get(episode.season) || { known: 0, aired: 0 };
      stats.known += 1;

      if (isEpisodeAired(episode, now, today)) {
        stats.aired += 1;
      }

      episodeStatsBySeason.set(episode.season, stats);
    });

    data.seasons.forEach((season) => {
      const stats = episodeStatsBySeason.get(season.number) || { known: 0, aired: 0 };
      const total =
        season.episodeOrder && season.episodeOrder > 0
          ? season.episodeOrder
          : stats.known;
      const aired = Math.min(stats.aired, total);

      progressBySeason.set(season.number, {
        aired,
        total,
        complete: total > 0 && aired >= total,
      });
    });

    return progressBySeason;
  }, [data]);

  const loadShowDetails = async () => {
    if (!showID || Number.isNaN(showID)) {
      setError("Invalid series id.");
      setLoading(false);
      return;
    }

    setLoading(true);
    try {
      const response = await axios.get(`${domain}/api/tvmaze/series/${showID}`, {
        withCredentials: true,
      });

      const payload: SeriesDetailResponse = response.data;
      setData(payload);
      const preferredQuality =
        payload.installStatus?.subscription?.preferredQuality || "1080";
      setQuality(preferredQuality);
      setAutoInstallUpcoming(
        payload.installStatus?.subscription?.autoInstallUpcoming === true
      );
      setError("");
    } catch (err: any) {
      if (err.response?.status === 401) {
        await authProvider.signout();
        navigate("/login");
        return;
      }

      setError(err.response?.data?.error || "Failed to load series details");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadShowDetails();
  }, [showID]);

  const withAction = async (actionKey: string, callback: () => Promise<void>) => {
    setWorkingAction(actionKey);
    setMessage("");
    setError("");

    try {
      await callback();
      await loadShowDetails();
    } catch (err: any) {
      if (err.response?.status === 401) {
        await authProvider.signout();
        navigate("/login");
        return;
      }

      setError(err.response?.data?.error || "Action failed");
    } finally {
      setWorkingAction("");
    }
  };

  const configureAutoInstall = async (enabled: boolean) => {
    await withAction("auto", async () => {
      await axios.put(
        `${domain}/api/tvmaze/series/${showID}/auto-install`,
        {
          enabled,
          quality,
        },
        {
          withCredentials: true,
        }
      );
      setAutoInstallUpcoming(enabled);
      setMessage(
        enabled
          ? "Auto-install for upcoming episodes enabled"
          : "Auto-install for upcoming episodes disabled"
      );
    });
  };

  const requestAutoInstallChange = (enabled: boolean) => {
    if (enabled === autoInstallUpcoming || workingAction === "auto") {
      return;
    }

    setPendingAutoInstallUpcoming(enabled);
    setShowAutoInstallModal(true);
  };

  const closeAutoInstallModal = () => {
    setShowAutoInstallModal(false);
    setPendingAutoInstallUpcoming(null);
  };

  const confirmAutoInstallChange = async () => {
    if (pendingAutoInstallUpcoming === null) {
      return;
    }

    const enabled = pendingAutoInstallUpcoming;
    closeAutoInstallModal();
    await configureAutoInstall(enabled);
  };

  const installWholeShow = async () => {
    await withAction("install-show", async () => {
      const response = await axios.post<QueueResult>(
        `${domain}/api/tvmaze/series/${showID}/install/show`,
        { quality },
        { withCredentials: true }
      );

      setMessage(
        response.data.queued > 0
          ? `Queued ${response.data.queued} episodes (${response.data.skipped} skipped).`
          : "All aired episodes are already installed for this quality."
      );
    });
  };

  const installSeason = async (seasonNumber: number) => {
    await withAction(`season-${seasonNumber}`, async () => {
      const response = await axios.post<QueueResult>(
        `${domain}/api/tvmaze/series/${showID}/install/season/${seasonNumber}`,
        { quality },
        { withCredentials: true }
      );

      setMessage(
        response.data.queued > 0
          ? `Queued ${response.data.queued} episodes from season ${seasonNumber}.`
          : `All aired episodes in season ${seasonNumber} are already installed for this quality.`
      );
    });
  };

  const installEpisode = async (episodeID: number) => {
    await withAction(`episode-${episodeID}`, async () => {
      const response = await axios.post<QueueResult>(
        `${domain}/api/tvmaze/series/${showID}/install/episode/${episodeID}`,
        { quality },
        { withCredentials: true }
      );

      setMessage(
        response.data.queued > 0
          ? "Episode queued for installation."
          : "Episode already installed for this quality."
      );
    });
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="skeleton h-64 w-full" />
        <div className="skeleton h-40 w-full" />
        <div className="skeleton h-64 w-full" />
      </div>
    );
  }

  if (!data) {
    return <div className="alert alert-error">{error || "Series not found."}</div>;
  }

  const scheduleDays = data.show.schedule?.days?.join(", ") || "Unknown days";
  const scheduleTime = data.show.schedule?.time || "Unknown time";
  const isShowEnded = (data.show.status || "").trim().toLowerCase() === "ended";
  const airedSeasons = data.seasons.filter((season) => {
    const progress = seasonProgressByNumber.get(season.number);
    return (progress?.aired || 0) > 0;
  });
  const now = Date.now();
  const today = new Date().toISOString().slice(0, 10);
  const airedEpisodes = data.episodes.filter((episode) =>
    isEpisodeAired(episode, now, today)
  );
  const airedEpisodeIDsBySeason = new Map<number, number[]>();
  airedEpisodes.forEach((episode) => {
    const seasonEpisodes = airedEpisodeIDsBySeason.get(episode.season) || [];
    seasonEpisodes.push(episode.id);
    airedEpisodeIDsBySeason.set(episode.season, seasonEpisodes);
  });
  const isWholeShowInstalled =
    airedEpisodes.length > 0 &&
    airedEpisodes.every((episode) => downloadedEpisodeIDsForQuality.has(episode.id));

  return (
    <div className="space-y-6">
      <div className="grid gap-5 rounded-xl border border-base-300 bg-base-200 p-4 md:grid-cols-[220px,1fr]">
        <img
          src={data.show.image?.original || data.show.image?.medium || FALLBACK_POSTER}
          alt={data.show.name}
          className="w-full rounded-xl object-cover"
          loading="lazy"
        />

        <div className="space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-2xl font-semibold">{data.show.name}</h1>
            <span className={`badge ${statusBadgeClass(data.show.status)}`}>
              {data.show.status || "Unknown"}
            </span>
          </div>

          <div className="text-sm opacity-80">
            <p>Premiered: {data.show.premiered || "Unknown"}</p>
            {!isShowEnded ? (
              <p>
                Schedule: {scheduleDays} at {scheduleTime}
              </p>
            ) : null}
          </div>

          {data.show.summary ? (
            <div
              className="prose prose-sm max-w-none"
              dangerouslySetInnerHTML={{ __html: data.show.summary }}
            />
          ) : null}

          <div className="flex flex-wrap gap-2">
            <span className="badge badge-outline">
              Pending: {data.installStatus?.pendingCount || 0}
            </span>
            <span className="badge badge-success">
              Downloaded: {data.installStatus?.downloadedCount || 0}
            </span>
            <span className="badge badge-error">
              Failed: {data.installStatus?.failedCount || 0}
            </span>
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-base-300 bg-base-200 p-4">
        <h2 className="mb-3 text-lg font-semibold">Install Settings</h2>

        <div className="flex flex-wrap items-center gap-4">
          <div className="flex items-center gap-2">
            <span className="label-text text-sm">Preferred Quality</span>
            <div className="join">
              <input
                className="join-item btn"
                type="radio"
                name="preferred-quality"
                aria-label="1080p"
                checked={quality === "1080"}
                onChange={() => setQuality("1080")}
              />
              <input
                className="join-item btn"
                type="radio"
                name="preferred-quality"
                aria-label="2160p"
                checked={quality === "2160"}
                onChange={() => setQuality("2160")}
              />
            </div>
          </div>

          {!isShowEnded ? (
            <label className="label cursor-pointer gap-2 p-0">
              <span className="label-text">Auto-install upcoming episodes</span>
              <input
                type="checkbox"
                className="checkbox checkbox-success"
                checked={autoInstallUpcoming}
                disabled={workingAction === "auto"}
                onChange={(event) => requestAutoInstallChange(event.target.checked)}
              />
            </label>
          ) : null}

          <button
            className="btn btn-outline"
            disabled={workingAction === "install-show" || isWholeShowInstalled}
            onClick={installWholeShow}
          >
            {workingAction === "install-show"
              ? "Queueing..."
              : isWholeShowInstalled
                ? "Installed"
                : "Install Whole Show"}
          </button>
        </div>
      </div>

      {showAutoInstallModal && !isShowEnded ? (
        <div className="modal modal-open" role="dialog" aria-modal="true">
          <div className="modal-box">
            <h3 className="text-lg font-semibold">
              {pendingAutoInstallUpcoming ? "Enable auto-install?" : "Disable auto-install?"}
            </h3>
            <p className="py-3 text-sm opacity-80">
              {pendingAutoInstallUpcoming
                ? `Upcoming episodes will be auto-installed at ${quality}p quality.`
                : "Upcoming episodes will no longer be auto-installed."}
            </p>
            <div className="modal-action">
              <button className="btn btn-ghost" onClick={closeAutoInstallModal}>
                Cancel
              </button>
              <button className="btn btn-primary" onClick={confirmAutoInstallChange}>
                {pendingAutoInstallUpcoming ? "Enable" : "Disable"}
              </button>
            </div>
          </div>
          <form
            method="dialog"
            className="modal-backdrop"
            onSubmit={(event) => {
              event.preventDefault();
              closeAutoInstallModal();
            }}
          >
            <button aria-label="Close auto-install modal">close</button>
          </form>
        </div>
      ) : null}

      {message ? <div className="alert alert-success">{message}</div> : null}
      {error ? <div className="alert alert-error">{error}</div> : null}

      <div className="rounded-xl border border-base-300 bg-base-200 p-4">
        <h2 className="mb-3 text-lg font-semibold">Seasons</h2>
        <div className="space-y-2">
          {airedSeasons.length === 0 ? (
            <div className="rounded-lg border border-base-300 bg-base-100 p-3 text-sm opacity-70">
              No aired seasons yet.
            </div>
          ) : null}
          {airedSeasons.map((season) => {
            const progress = seasonProgressByNumber.get(season.number) || {
              aired: 0,
              total: season.episodeOrder || 0,
              complete: false,
            };
            const seasonEpisodeIDs = airedEpisodeIDsBySeason.get(season.number) || [];
            const isSeasonInstalled =
              seasonEpisodeIDs.length > 0 &&
              seasonEpisodeIDs.every((episodeID) =>
                downloadedEpisodeIDsForQuality.has(episodeID)
              );

            return (
              <div
                key={season.id}
                className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-base-300 bg-base-100 p-3"
              >
                <div>
                  <div className="flex items-center gap-2">
                    <p className="font-medium">Season {season.number}</p>
                    <span
                      className={`badge badge-sm ${progress.complete ? "badge-success" : "badge-warning"}`}
                    >
                      {progress.complete ? "Complete" : "Incomplete"}
                    </span>
                  </div>
                  <p className="text-xs opacity-70">
                    {season.premiereDate || "Unknown"} - {season.endDate || "Unknown"} - Aired:{" "}
                    {progress.aired}/{progress.total}
                  </p>
                </div>
                <button
                  className="btn btn-sm btn-outline"
                  disabled={workingAction === `season-${season.number}` || isSeasonInstalled}
                  onClick={() => installSeason(season.number)}
                >
                  {workingAction === `season-${season.number}`
                    ? "Queueing..."
                    : isSeasonInstalled
                      ? "Installed"
                      : "Install Season"}
                </button>
              </div>
            );
          })}
        </div>
      </div>

      <div className="rounded-xl border border-base-300 bg-base-200 p-4">
        <h2 className="mb-3 text-lg font-semibold">Episodes</h2>

        {airedEpisodes.length === 0 ? (
          <div className="rounded-lg border border-base-300 bg-base-100 p-3 text-sm opacity-70">
            No aired episodes yet.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="table table-zebra table-sm">
              <thead>
                <tr>
                  <th>S/E</th>
                  <th>Title</th>
                  <th>Airs</th>
                  <th>Status</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {airedEpisodes.map((episode) => {
                  const job = jobsByEpisodeID.get(episode.id);
                  const status = job?.status || "not queued";
                  const isEpisodeInstalled = downloadedEpisodeIDsForQuality.has(episode.id);

                  return (
                    <tr key={episode.id}>
                      <td>
                        S{episode.season}E{episode.number}
                      </td>
                      <td>{episode.name}</td>
                      <td>
                        {episode.airdate || "Unknown"} {episode.airtime || ""}
                      </td>
                      <td>
                        <span className={`badge badge-xs ${jobBadgeClass(status)}`}>
                          {status}
                        </span>
                      </td>
                      <td>
                        <button
                          className="btn btn-xs btn-outline"
                          disabled={workingAction === `episode-${episode.id}` || isEpisodeInstalled}
                          onClick={() => installEpisode(episode.id)}
                        >
                          {workingAction === `episode-${episode.id}`
                            ? "Queueing..."
                            : isEpisodeInstalled
                              ? "Installed"
                              : "Install"}
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
