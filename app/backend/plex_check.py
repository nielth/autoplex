import asyncio
import requests
import xml.etree.ElementTree as ET

from plexapi.server import PlexServer
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

PLEX_API_BASE_URL='https://plex.tv/users/account/?X-Plex-Token='


def check_plex_user(token):
    xml = requests.get(PLEX_API_BASE_URL + token)
    xml_parsed = ET.fromstring(xml.content)
    email = xml_parsed.attrib['email']
    id = xml_parsed.attrib['id']
    uuid = xml_parsed.attrib['uuid']
    username = xml_parsed.attrib['username']
    return email


def redirect_plex(plexauth):
    return f'<a href"{plexauth.auth_url(forward_url="http://localhost:1337")} target="_blank">Link</a>"'


