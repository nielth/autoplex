import axios from "axios";
import { formatBytes } from "../scripts/formatBytes";
import {
  ButtonDownloadIcon,
  CompletedIcon,
  LeechersIcon,
  SeedersIcon,
  SizeIcon,
  TimestampIcon,
} from "./Icons";
import { getApiDomain } from "../scripts/getApiDomain";

function torrent_download(data: {
  fid: string;
  filename: string;
  categoryID: string;
}) {
  const domain = getApiDomain();
  axios.post(
    `${domain}/api/download`,
    {
      fid: data.fid,
      filename: data.filename,
      categoryID: data.categoryID,
    },
    { withCredentials: true }
  );
}

export function TorrentList({ data }: { data: TorrentData }) {
  return (
    <>
      <div className="divider my-2"></div>
      {data.torrentList.map((torrent) => (
        <div key={torrent.fid}>
          <div className="flex gap-x-1">
            <div className="grow">
              <div className="flex items-center gap-x-2">
                <p className="md:text-base text-xs">{torrent.name}</p>
              </div>
              <div className="flex lg:gap-x-3 gap-x-2 md:pt-0 pt-2 text-xs">
                <div className="flex lg:gap-x-3 gap-x-2 text-xs">
                  <div
                    className="tooltip tooltip-bottom"
                    data-tip="Torrent uploaded"
                  >
                    <p className="hidden md:flex gap-x-1 items-center opacity-60">
                      <TimestampIcon width={"16"} height={"16"} />
                      {torrent.addedTimestamp}
                    </p>
                  </div>
                  <div className="tooltip tooltip-bottom" data-tip="Filesize">
                    <p className="flex gap-x-1 items-center opacity-60">
                      <SizeIcon width={"16"} height={"16"} />
                      {formatBytes(torrent.size)}
                    </p>
                  </div>
                  <div
                    className="tooltip tooltip-bottom"
                    data-tip="Number downloads"
                  >
                    <p className="hidden md:flex gap-x-1 items-center opacity-60">
                      <CompletedIcon height={"16"} width={"16"} />
                      {torrent.completed}
                    </p>
                  </div>
                  <div className="tooltip tooltip-bottom" data-tip="Seeders">
                    <p className="flex gap-x-1 items-center text-green-600 opacity-60">
                      <SeedersIcon height={"16"} width={"16"} />
                      {torrent.seeders}
                    </p>
                  </div>
                  <div className="tooltip tooltip-bottom" data-tip="Leechers">
                    <p className="hidden md:flex gap-x-1 items-center text-red-600 opacity-60">
                      <LeechersIcon height={"16"} width={"16"} />
                      {torrent.leechers}
                    </p>
                  </div>
                </div>
                {torrent.tags.includes("FREELEECH") ? (
                  <div className="bg-[#ffdf00] rounded px-0.5 py-0.5">
                    <p className="text-[#4d4d4d] text-xs">FREELEECH</p>
                  </div>
                ) : null}
              </div>
            </div>
            <button
              className="btn btn-ghost text-[#e5a00d]"
              onClick={() => {
                torrent_download({
                  fid: torrent.fid,
                  filename: torrent.filename,
                  categoryID: String(torrent.categoryID),
                });
              }}
            >
              <ButtonDownloadIcon height={"20"} width={"20"} />
            </button>
          </div>
          <div className="divider my-2"></div>
        </div>
      ))}
    </>
  );
}
