import os
import hashlib
import bencodepy

from qbittorrent import Client

USER = os.getenv("QBT_USER")
PASSWORD = os.getenv("QBT_PASS")
QBT_URL = os.getenv("QBT_URL")


def get_info_hash_v1(torrent_file):
    torrent_data = bencodepy.decode(torrent_file)
    info = torrent_data[b"info"]
    info_hash = hashlib.sha1(bencodepy.encode(info)).hexdigest()
    return info_hash


def authenticate():
    qb = Client(QBT_URL, verify=False)
    qb.login(USER, PASSWORD)
    return qb


def qbtTorrentDownload(torrent_file, category):
    qb = authenticate()
    dl_path = f"/downloads/sdb/{category}"
    qb.download_from_file(torrent_file, savepath=dl_path)
    torrent_hash = get_info_hash_v1(torrent_file)
    qb.toggle_sequential_download(torrent_hash)
