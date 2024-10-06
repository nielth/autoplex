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
import { getCookie } from "../scripts/getCookie";
import { useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";

async function torrent_download(
  data: {
    fid: string;
    filename: string;
    categoryID: string;
  },
  navigate: Function,
  setAlertStatus: Function
) {
  const domain = getApiDomain();
  setAlertStatus(false);
  axios
    .post(
      `${domain}/api/download`,
      {
        fid: data.fid,
        filename: data.filename,
        categoryID: data.categoryID,
      },
      {
        withCredentials: true,
        headers: {
          "X-CSRF-TOKEN": getCookie("csrf_access_token"),
        },
      }
    )
    .then((resp) => {
      console.log(resp);
      setAlertStatus(true);
    })
    .catch((error) => {
      if (error.status === 500) {
        navigate("/error");
      }
    });
}

export function TorrentList({ data }: { data: TorrentData }) {
  const [alertStatus, setAlertStatus] = useState<boolean>(false);
  const navigate = useNavigate();

  useEffect(() => {
    if (alertStatus === true) {
      setTimeout(() => {
        setAlertStatus(false);
      }, 5000);
    }
  }, [alertStatus]);

  return (
    <>
      {alertStatus ? (
        <div className="fixed left-1/2 md:top-5 bottom-5 md:h-0">
          <div
            role="alert"
            className="alert alert-success min-w-max relative left-[-50%]"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-6 w-6 shrink-0 stroke-current"
              fill="none"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>Your purchase has been confirmed!</span>
          </div>
        </div>
      ) : null}
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
                torrent_download(
                  {
                    fid: torrent.fid,
                    filename: torrent.filename,
                    categoryID: String(torrent.categoryID),
                  },
                  navigate,
                  setAlertStatus
                );
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
