import { useEffect, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { TorrentList } from "../components/TorrentList";
import axios from "axios";
import { authProvider } from "../auth";

async function search_torrent(
  search: string,
  page: number,
  navigate: Function,
  setData: Function,
  setDataLoading: Function
) {
  await setDataLoading(true);
  const answer = await axios
    .get(`http://localhost:5050/api/search/${search}/${page}`, {
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
  await setDataLoading(false);
  return answer;
}

export function Home() {
  const [search, setSearch] = useState<string>("");
  const [data, setData] = useState<TorrentData>(Object);
  const [page, setPage] = useState<number>(1);
  const [dataLoading, setDataLoading] = useState<Boolean>(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();

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

  const handleSearch = (event: any) => {
    if (event.key === "Enter" || event.type === "click") {
      setSearchParams(`search=${search}&p=1`);
      setPage(1);
      search_torrent(search, 1, navigate, setData, setDataLoading);
    }
  };

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

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
    }
  });

  useEffect(() => {
    console.log(data);
  }, [data]);

  return (
    <>
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
              {[...Array(35).keys()].map((key) => (
                <div>
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
