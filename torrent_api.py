import qbittorrentapi

# instantiate a Client using the appropriate WebUI configuration
qbt_client = qbittorrentapi.Client(
    host='192.168.10.112',
    port=8081,
    username='admin',
    password='adminadmin',
)


def add_movie_to_api(magnet):
    qbt_client.torrents_add(magnet)
