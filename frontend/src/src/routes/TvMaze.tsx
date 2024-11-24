import axios from "axios";
import { getApiDomain } from "../scripts/getApiDomain";
import { useState } from "react";

interface TvMazeSearch {
  averageRuntime: number;
  image: {
    medium: string;
    original: string;
  };
}

export function TvMaze() {
  const domain = getApiDomain();
  const [search, setSearch] = useState<String>("");
  const [data, setData] = useState<TvMazeSearch>(Object());

  const handleSearch = (event: any) => {
    if (event.key === "Enter" || event.type === "click") {
      axios
        .get(`${domain}/api/tvmaze/search/${search}`, { withCredentials: true })
        .then((resp) => {
          if (resp.status === 200) {
            console.log(resp.data);
            setData(resp.data);
          }
        });
    }
  };

  return (
    <>
      <div>
        <input
          type="text"
          placeholder="Type here"
          onChange={(e) => {
            setSearch(e.target.value);
          }}
          onKeyDown={handleSearch}
          className="input input-bordered w-full max-w-xs"
        />
      </div>
      {data ? (
        <>{<img src={data.image.medium} alt="tvmazesearchimage" />}</>
      ) : null}
    </>
  );
}
