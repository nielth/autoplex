import axios from "axios";
import { FormEvent, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";

import { authProvider } from "../auth";
import { getApiDomain } from "../scripts/getApiDomain";

interface TvMazeImage {
  medium?: string;
  original?: string;
}

interface TvMazeShow {
  id: number;
  name: string;
  status?: string;
  premiered?: string;
  image?: TvMazeImage;
}

interface TvMazeSearchResult {
  score: number;
  show: TvMazeShow;
}

const FALLBACK_POSTER =
  "https://static.tvmaze.com/images/no-img/no-img-portrait-text.png";

export function TvMaze() {
  const [query, setQuery] = useState<string>("");
  const [searchedQuery, setSearchedQuery] = useState<string>("");
  const [results, setResults] = useState<TvMazeSearchResult[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");
  const navigate = useNavigate();
  const domain = getApiDomain();

  const canSearch = useMemo(() => query.trim().length > 1, [query]);

  const handleSearch = async (event: FormEvent) => {
    event.preventDefault();

    const cleanQuery = query.trim();
    if (cleanQuery.length < 2) {
      setResults([]);
      setError("Enter at least 2 characters.");
      return;
    }

    setLoading(true);
    setError("");
    setSearchedQuery(cleanQuery);

    try {
      const response = await axios.get(`${domain}/api/tvmaze/search`, {
        withCredentials: true,
        params: { q: cleanQuery },
      });

      setResults(response.data?.results ?? []);
    } catch (err: any) {
      if (err.response?.status === 401) {
        await authProvider.signout();
        navigate("/login");
        return;
      }

      setError(err.response?.data?.error || "Failed to search TVMaze");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">TV Series Scheduler</h1>
        <p className="text-sm opacity-70">
          Search TVMaze series, then configure auto-install rules and quality.
        </p>
      </div>

      <form className="flex flex-wrap gap-2" onSubmit={handleSearch}>
        <input
          type="text"
          className="input input-bordered w-full md:max-w-xl"
          placeholder="Search TV series"
          autoFocus
          value={query}
          onChange={(event) => setQuery(event.target.value)}
        />
        <button className="btn btn-primary" disabled={!canSearch || loading}>
          {loading ? "Searching..." : "Search"}
        </button>
      </form>

      {error ? <div className="alert alert-error">{error}</div> : null}

      {!loading &&
      !error &&
      results.length === 0 &&
      searchedQuery.length > 0 &&
      query.trim() === searchedQuery ? (
        <div className="rounded-xl border border-base-300 bg-base-200 p-5 text-sm opacity-80">
          No matching series found.
        </div>
      ) : null}

      {loading ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
          {[...Array(10).keys()].map((value) => (
            <div key={value} className="skeleton h-72 w-full" />
          ))}
        </div>
      ) : null}

      {!loading && results.length > 0 ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
          {results.map((result) => (
            <Link
              to={`/series/${result.show.id}`}
              key={`${result.show.id}-${result.score}`}
              className="card border border-base-300 bg-base-200 transition hover:border-primary hover:shadow-lg"
            >
              <figure className="aspect-[2/3] overflow-hidden bg-base-300">
                <img
                  src={result.show.image?.original || result.show.image?.medium || FALLBACK_POSTER}
                  alt={result.show.name}
                  className="h-full w-full object-cover"
                  loading="lazy"
                />
              </figure>
              <div className="card-body p-3">
                <p className="line-clamp-2 text-sm font-medium">{result.show.name}</p>
                <p className="text-xs opacity-70">
                  {result.show.premiered || "Unknown date"} - {result.show.status || "Unknown"}
                </p>
              </div>
            </Link>
          ))}
        </div>
      ) : null}
    </div>
  );
}
