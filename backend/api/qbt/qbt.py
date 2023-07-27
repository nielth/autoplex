import os

from qbittorrent import Client

USER = os.getenv('QBT_USER')
PASSWORD = os.getenv('QBT_PASS')


def authenticate():
    qb = Client('http://192.168.0.165:8080/')
    qb.login(USER, PASSWORD)
    return qb


def qbtTorrentDownload(torrent_file):
    qb = authenticate()
    dl_path = '/downloads'
    qb.download_from_file(torrent_file, savepath=dl_path)
