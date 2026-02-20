import { useEffect, useRef, useState } from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { TorrentList } from "../components/TorrentList";
import axios from "axios";
import { authProvider } from "../auth";
import { getApiDomain } from "../scripts/getApiDomain";
import { formatBytes } from "../scripts/formatBytes";

interface ActiveDownload {
  hash: string;
  name: string;
  progress: number;
  size: number;
  num_seeds: number;
  num_leechs: number;
}

async function search_torrent(
  search: string,
  page: number,
  navigate: Function,
  setData: Function,
  setDataLoading: Function
) {
  setDataLoading(true);
  const domain = await getApiDomain();
  const answer = await axios
    .get(`${domain}/api/search/${search}/${page}`, {
      withCredentials: true,
    })
    .then((resp) => {
      if (resp.status && resp.status === 200) {
        setData(resp.data);
      }
    })
    .catch((error) => {
      console.log(error);
      if (error.status === 401) {
        authProvider.signout();
        navigate("/login");
      } else {
        navigate("/error");
      }
    });
  setDataLoading(false);
  return answer;
}

export function Home() {
  const [search, setSearch] = useState<string>("");
  const [data, setData] = useState<TorrentData>(Object);
  const [page, setPage] = useState<number>(1);
  const [dataLoading, setDataLoading] = useState<boolean>(false);
  const [activeDownloads, setActiveDownloads] = useState<ActiveDownload[]>([]);
  const inputRef = useRef<HTMLInputElement>(null);
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();
  const domain = getApiDomain();

  async function pagination(next_or_prev: boolean, old_page: number) {
    setData(Object());

    if (
      next_or_prev == true &&
      old_page < Math.ceil(data.numFound / data.perPage)
    ) {
      setPage(old_page + 1);
      setSearchParams(`search=${search}&p=${old_page + 1}`);
    } else if (next_or_prev == false && page > 1) {
      setPage(old_page - 1);
      setSearchParams(`search=${search}&p=${old_page - 1}`);
    }
  }

  // If search is initiated with "Enter" or "click", search the text input
  const handleSearch = (event: any) => {
    if (event.key === "Enter" || event.type === "click") {
      setPage(1);
      const currentSearch = searchParams.get("search") ?? "";
      const currentPage = searchParams.get("p") ?? "";

      if (currentSearch === search && currentPage === "1") {
        search_torrent(search, 1, navigate, setData, setDataLoading);
        return;
      }

      setSearchParams(`search=${search}&p=1`);
    }
  };

  // Grabs search parameters every change, and search the corresponding data
  useEffect(() => {
    const sparmSearch = searchParams.get("search");
    const sparmPage = searchParams.get("p");

    if (sparmSearch !== null && sparmPage !== null) {
      setSearch(sparmSearch);
      setPage(Number(sparmPage));
      search_torrent(
        sparmSearch,
        Number(sparmPage),
        navigate,
        setData,
        setDataLoading
      );
    } else {
      setSearch("");
      setPage(1);
      setData(Object());
    }
  }, [searchParams]);

  // Auto focus text input on load (not mobile)
  useEffect(() => {
    if (inputRef.current && window.innerWidth >= 900) {
      inputRef.current.focus();
    }
  }, []);

  useEffect(() => {
    setData(Object());
  }, [location]);

  useEffect(() => {
    const fetchActiveDownloads = () => {
      axios
        .get(`${domain}/api/disk`, { withCredentials: true })
        .then((resp) => {
          setActiveDownloads(resp.data.qbtDownloadingList ?? []);
        })
        .catch((error) => {
          if (error.response?.status === 401) {
            authProvider.signout();
            navigate("/login");
          }
        });
    };

    fetchActiveDownloads();
    const intervalId = window.setInterval(fetchActiveDownloads, 10000);

    return () => {
      window.clearInterval(intervalId);
    };
  }, [domain, navigate]);

  return (
    <>
      {activeDownloads.length > 0 ? (
        <div className="mb-8 rounded-xl border border-base-300 bg-base-200 p-4">
          <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide opacity-70">
            Active Downloads
          </h2>
          <div className="space-y-3">
            {activeDownloads.map((download) => {
              const progressPercent = Math.round(download.progress * 10000) / 100;

              return (
                <div key={download.hash || download.name} className="space-y-1">
                  <div className="flex flex-col gap-1 text-sm sm:flex-row sm:items-center sm:justify-between">
                    <p className="font-medium break-all">{download.name}</p>
                    <p className="opacity-70">
                      {formatBytes(download.size)} - {progressPercent}%
                    </p>
                  </div>
                  <progress
                    className="progress progress-info w-full"
                    value={progressPercent}
                    max={100}
                  />
                </div>
              );
            })}
          </div>
        </div>
      ) : null}
      <div className="flex justify-center gap-x-1">
        <input
          ref={inputRef}
          type="text"
          placeholder="Search TV or Movie"
          onChange={(e) => {
            setSearch(e.target.value);
          }}
          onKeyDown={handleSearch}
          value={search}
          className="input input-bordered w-full max-w-md"
        />
        <button
          onClick={handleSearch}
          className="hidden md:block btn btn-outline"
        >
          Search
        </button>
      </div>
      <div>
        <div className="pt-8">
          {dataLoading ? (
            <div>
              {[...Array(35).keys()].map((_, index) => (
                <div key={index}>
                  <div className="divider my-2" />
                  <div className="skeleton h-12 w-full"></div>
                </div>
              ))}
            </div>
          ) : null}

          {data.torrentList ? (
            <div className="mx-auto">
              <TorrentList data={data} />
              <div className="text-center">
                <div className="join">
                  {page <= 1 ? (
                    <button className="join-item btn btn-disabled">«</button>
                  ) : (
                    <button
                      className="join-item btn"
                      onClick={() => pagination(false, page)}
                    >
                      «
                    </button>
                  )}
                  <button className="join-item btn">
                    Page {page}/{Math.ceil(data.numFound / data.perPage)}
                  </button>
                  {page >= Math.ceil(data.numFound / data.perPage) ? (
                    <button className="join-item btn btn-disabled">»</button>
                  ) : (
                    <button
                      className="join-item btn"
                      onClick={() => pagination(true, page)}
                    >
                      »
                    </button>
                  )}
                </div>
              </div>
            </div>
          ) : null}
        </div>
      </div>
    </>
  );
}
