import qbittorrentapi

# instantiate a Client using the appropriate WebUI configuration
qbt_client = qbittorrentapi.Client(
    host='10.0.0.33',
    port=8081,
    username='admin',
    password='adminadmin',
)


def torrentAPI(magnet, category):
    if category == "movie":
        qbt_client.torrents_add(magnet, save_path='/movies')
    elif category == "series":
        qbt_client.torrents_add(magnet, save_path='/tvseries')
