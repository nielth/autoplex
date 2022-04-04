from dotenv import load_dotenv

import threading
import qbittorrentapi
import time
import os

load_dotenv()

qbt_client = qbittorrentapi.Client(
    host="172.17.0.1", port="8080", username="admin", password="adminadmin"
)

qbt_client_now = qbittorrentapi.Client(
    host="172.17.0.1", port="8081", username="admin", password="adminadmin"
)


def torrent_api(magnet, category):
    if category == "movie":
        qbt_client.torrents_add(
            magnet, save_path="/downloads/movies", is_sequential_download=True
        )
    elif category == "series":
        qbt_client.torrents_add(
            magnet, save_path="/downloads/tvseries", is_sequential_download=True
        )


def download_status():
    torrent_list_active = qbt_client.torrents.info.downloading()
    torrent_list_active_norw = qbt_client_now.torrents.info.downloading()
    return_title = []
    return_status = []
    for torrent in torrent_list_active:
        return_title.append(torrent["name"])
        return_status.append(round((float(torrent["progress"]) * 100), 1))
    for torrent in torrent_list_active_norw:
        return_title.append(torrent["name"])
        return_status.append(round((float(torrent["progress"]) * 100), 1))

    return return_title, return_status
