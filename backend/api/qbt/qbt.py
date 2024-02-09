import os

from qbittorrent import Client

USER = os.getenv("QBT_USER")
PASSWORD = os.getenv("QBT_PASS")


def authenticate():
    qb = Client("https://qbt.internal.nielth.com/", verify=False)
    qb.login(USER, PASSWORD)
    return qb


def qbtTorrentDownload(torrent_file):
    qb = authenticate()
    dl_path = "/downloads/sdb/tvseries"
    qb.download_from_file(torrent_file, savepath=dl_path)
