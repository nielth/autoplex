from plexapi.server import PlexServer

def check_plex_user():
    token = '2ffLuB84dqLswk9skLos'
    baseurl = 'http://plex:32400'
    plex = PlexServer(baseurl, token)