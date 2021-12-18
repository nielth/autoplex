import qbittorrentapi

# instantiate a Client using the appropriate WebUI configuration
qbt_client = qbittorrentapi.Client(
    host='localhost',
    port=8081,
    username='thomasnie',
    password='userpass',
)


def add_movie_to_api(magnet):
    qbt_client.torrents_add(magnet)
