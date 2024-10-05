import axios from "axios";
import { useEffect, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";

export function Home() {
  const [search, setSearch] = useState<string>("");
  const [page, setPage] = useState<number>(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const [searchParams, setSearchParams] = useSearchParams();

  const handleSearch = (event: any) => {
    if (event.key === "Enter" || event.type === "click") {
      setSearchParams(`search=${search}&p=0`);
      axios
        .get(
          `https://autoplex.nielth.com/api/search/${search}/${
            Number(page) + 1
          }`,
          {
            withCredentials: true,
          }
        )
        .then((resp: any) => {
          console.log(resp.content);
        });
    }
  };

  useEffect(() => {
    const sparmSearch = searchParams.get("search");
    const sparmPage = searchParams.get("p");

    if (sparmSearch !== null && sparmPage !== null) {
      setSearch(sparmSearch);
      setPage(Number(sparmPage));
    } else {
      setSearch("");
      setPage(0);
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
    </>
  );
}
