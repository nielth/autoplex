import { formatBytes } from "../scripts/formatBytes";

export function TorrentList({ data }: { data: TorrentData }) {
  return (
    <>
      <div className="pt-8">
        <div className="divider my-2"></div>
        {data.torrentList.map((torrent) => (
          <div id={torrent.fid}>
            <div className="flex gap-x-1">
              <div className="grow">
                <div className="flex items-center gap-x-2">
                  <p className="md:text-base text-xs">{torrent.name}</p>
                </div>
                <div className="flex lg:gap-x-3 gap-x-2 md:pt-0 pt-2 text-xs">
                  <div className="flex lg:gap-x-3 gap-x-2 text-xs opacity-60">
                    <p className="hidden lg:flex gap-x-1 items-center">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="16"
                        height="16"
                        fill="currentColor"
                        className="bi bi-clock"
                        viewBox="0 0 16 16"
                      >
                        <path d="M8 3.5a.5.5 0 0 0-1 0V9a.5.5 0 0 0 .252.434l3.5 2a.5.5 0 0 0 .496-.868L8 8.71z" />
                        <path d="M8 16A8 8 0 1 0 8 0a8 8 0 0 0 0 16m7-8A7 7 0 1 1 1 8a7 7 0 0 1 14 0" />
                      </svg>
                      {torrent.addedTimestamp}
                    </p>
                    <p className="flex gap-x-1 items-center">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="16"
                        height="16"
                        fill="currentColor"
                        className="bi bi-files"
                        viewBox="0 0 16 16"
                      >
                        <path d="M13 0H6a2 2 0 0 0-2 2 2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h7a2 2 0 0 0 2-2 2 2 0 0 0 2-2V2a2 2 0 0 0-2-2m0 13V4a2 2 0 0 0-2-2H5a1 1 0 0 1 1-1h7a1 1 0 0 1 1 1v10a1 1 0 0 1-1 1M3 4a1 1 0 0 1 1-1h7a1 1 0 0 1 1 1v10a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1z" />
                      </svg>
                      {formatBytes(torrent.size)}
                    </p>
                    <p className="hidden lg:flex gap-x-1 items-center">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="16"
                        height="16"
                        fill="currentColor"
                        className="bi bi-download"
                        viewBox="0 0 16 16"
                      >
                        <path d="M.5 9.9a.5.5 0 0 1 .5.5v2.5a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-2.5a.5.5 0 0 1 1 0v2.5a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2v-2.5a.5.5 0 0 1 .5-.5" />
                        <path d="M7.646 11.854a.5.5 0 0 0 .708 0l3-3a.5.5 0 0 0-.708-.708L8.5 10.293V1.5a.5.5 0 0 0-1 0v8.793L5.354 8.146a.5.5 0 1 0-.708.708z" />
                      </svg>
                      {torrent.completed}
                    </p>
                    <p className="flex gap-x-1 items-center text-green-600">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="16"
                        height="16"
                        fill="currentColor"
                        className="bi bi-caret-up"
                        viewBox="0 0 16 16"
                      >
                        <path d="M3.204 11h9.592L8 5.519zm-.753-.659 4.796-5.48a1 1 0 0 1 1.506 0l4.796 5.48c.566.647.106 1.659-.753 1.659H3.204a1 1 0 0 1-.753-1.659" />
                      </svg>
                      {torrent.seeders}
                    </p>
                    <p className="hidden lg:flex gap-x-1 items-center text-red-600">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="16"
                        height="16"
                        fill="currentColor"
                        className="bi bi-caret-down"
                        viewBox="0 0 16 16"
                      >
                        <path d="M3.204 5h9.592L8 10.481zm-.753.659 4.796 5.48a1 1 0 0 0 1.506 0l4.796-5.48c.566-.647.106-1.659-.753-1.659H3.204a1 1 0 0 0-.753 1.659" />
                      </svg>
                      {torrent.leechers}
                    </p>
                  </div>
                  {torrent.tags.includes("FREELEECH") ? (
                    <div className="bg-[#ffdf00] rounded px-0.5 py-0.5">
                      <p className="text-[#4d4d4d] text-xs">FREELEECH</p>
                    </div>
                  ) : null}
                </div>
              </div>
              <button className="btn btn-ghost text-[#e5a00d]">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="20"
                  height="20"
                  fill="currentColor"
                  className="bi bi-download"
                  viewBox="0 0 16 16"
                >
                  <path d="M.5 9.9a.5.5 0 0 1 .5.5v2.5a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-2.5a.5.5 0 0 1 1 0v2.5a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2v-2.5a.5.5 0 0 1 .5-.5" />
                  <path d="M7.646 11.854a.5.5 0 0 0 .708 0l3-3a.5.5 0 0 0-.708-.708L8.5 10.293V1.5a.5.5 0 0 0-1 0v8.793L5.354 8.146a.5.5 0 1 0-.708.708z" />
                </svg>
              </button>
            </div>
            <div className="divider my-2"></div>
          </div>
        ))}
      </div>
    </>
  );
}
