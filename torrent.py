from dotenv import load_dotenv

import threading
import qbittorrentapi
import time
import os

load_dotenv()

# instantiate a Client using the appropriate WebUI configuration
qbt_client = qbittorrentapi.Client(
    host=os.getenv('ADDRESS'),
    port=os.getenv('PORT'),
    username=os.getenv('QBITTORRENT_USER'),
    password=os.getenv('QBITTORRENT_PASS'),
)

lock = threading.Lock()


def torrent_api(magnet, category):
    if category == "movie":
        qbt_client.torrents_add(
            magnet, save_path=os.getenv('MOVIE_PATH'), is_sequential_download=True
        )
    elif category == "series":
        qbt_client.torrents_add(
            magnet, save_path=os.getenv('TV_PATH'), is_sequential_download=True
        )


def delete_finished():
    while True:
        time.sleep(30)
        #with lock:
            #for torrent in qbt_client.torrents_info():
                # check if torrent is downloading
                #if torrent.state_enum.is_complete:
                #    qbt_client.torrents_delete(
                #        delete_files=False, torrent_hashes=torrent.hash
                #    )