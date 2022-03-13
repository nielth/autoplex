import asyncio
import requests
import xml.etree.ElementTree as ET

from plexapi.server import PlexServer
from plexauth import PlexAuth


PLEX_API_BASE_URL='https://plex.tv/users/account/?X-Plex-Token='


def check_plex_user(token):
    xml = requests.get(PLEX_API_BASE_URL + token)
    xml_parsed = ET.fromstring(xml.content)
    email = xml_parsed.attrib['email']
    id = xml_parsed.attrib['id']
    uuid = xml_parsed.attrib['uuid']
    username = xml_parsed.attrib['username']
    return email


import asyncio
from plexauth import PlexAuth

PAYLOAD = {
    'X-Plex-Product': 'Test Product',
    'X-Plex-Version': '0.0.1',
    'X-Plex-Device': 'Test Device',
    'X-Plex-Platform': 'Test Platform',
    'X-Plex-Device-Name': 'Test Device Name',
    'X-Plex-Device-Vendor': 'Test Vendor',
    'X-Plex-Model': 'Test Model',
    'X-Plex-Client-Platform': 'Test Client Platform'
}

async def temp():
    async with PlexAuth(PAYLOAD) as plexauth:
        await plexauth.initiate_auth()
        print("Complete auth at URL: {}".format(plexauth.auth_url()))
        token = await plexauth.token()
    
    if token:
        print("Token: {}".format(token))
    else:
        print("No token returned.")

loop = asyncio.get_event_loop()
loop.run_until_complete(temp())

