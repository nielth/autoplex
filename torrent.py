import threading
import qbittorrentapi
import time


# instantiate a Client using the appropriate WebUI configuration
qbt_client = qbittorrentapi.Client(
    host='10.0.0.33',
    port=8080,
    username='admin',
    password='adminadmin',
)

lock = threading.Lock()


def torrentAPI(magnet, category):
    if category == "movie":
        qbt_client.torrents_add(
            magnet, save_path='/downloads/movies', is_sequential_download=True)
    elif category == "series":
        qbt_client.torrents_add(
            magnet, save_path='/downloads/tvseries', is_sequential_download=True)


def deleteFinished():
    while(True):
        time.sleep(30)
        with lock:
            for torrent in qbt_client.torrents_info():
                # check if torrent is downloading
                if torrent.state_enum.is_complete:
                    qbt_client.torrents_delete(
                        delete_files=False, torrent_hashes=torrent.hash)
