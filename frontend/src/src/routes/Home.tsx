import { useEffect, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { temp } from "./temp";
import { TorrentList } from "../components/TorrentList";

export function Home() {
  const [search, setSearch] = useState<string>("");
  const [data, setData] = useState<TorrentData>(Object);
  const [page, setPage] = useState<number>(1);
  const inputRef = useRef<HTMLInputElement>(null);
  const [searchParams, setSearchParams] = useSearchParams();

  async function pagination(next_or_prev: boolean, old_page: number) {
    console.log(old_page);
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
    }
  };

  useEffect(() => {
    const sparmSearch = searchParams.get("search");
    const sparmPage = searchParams.get("p");
    setData(temp);
    console.log(temp);

    if (sparmSearch !== null && sparmPage !== null) {
      setSearch(sparmSearch);
      setPage(Number(sparmPage));
    } else {
      setSearch("");
      setPage(1);
    }
  }, [searchParams]);

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
    }
  });

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
        <button onClick={handleSearch} className="btn btn-outline">
          Search
        </button>
      </div>
      <div>
        {data && data.torrentList ? (
          <div className="mx-auto text-center">
            <TorrentList data={data} />
            <div className="join">
              {page <= 1 ? (
                <button
                  className="join-item btn btn-disabled"
                  onClick={() => pagination(false, page)}
                >
                  «
                </button>
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
                <button
                  className="join-item btn btn-disabled"
                  onClick={() => pagination(true, page)}
                >
                  »
                </button>
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
        ) : null}
      </div>
    </>
  );
}
