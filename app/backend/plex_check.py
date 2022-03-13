import requests
import xml.etree.ElementTree as ET
import time
import requests
import urllib.parse
import uuid

CODES_URL = 'https://plex.tv/api/v2/pins.json?strong=true'
AUTH_URL = 'https://app.plex.tv/auth#!?{}'
TOKEN_URL = 'https://plex.tv/api/v2/pins/{}'

PAYLOAD = {
    'X-Plex-Product': 'Test Product',
    'X-Plex-Version': '0.0.1',
    'X-Plex-Device': 'Test Device',
    'X-Plex-Platform': 'Test Platform',
    'X-Plex-Device-Name': 'Test Device Name',
    'X-Plex-Device-Vendor': 'Test Vendor',
    'X-Plex-Model': 'Test Model',
    'X-Plex-Client-Platform': 'Test Client Platform',
    'X-Plex-Client-Identifier': str(uuid.uuid4())
}

identifier = int
parameters = {
    'clientID': PAYLOAD['X-Plex-Client-Identifier'],
    'code': ''
}

s = None

def get_plex_link(forward_url=None):
    global s
    global identifier
    s = requests.Session()
    resp = s.post(CODES_URL, data=PAYLOAD, headers=None)
    response = resp.json()
    parameters['code'] = response['code']
    identifier = response['id']
    if forward_url:
        parameters['forwardUrl'] = forward_url
    url = AUTH_URL.format(urllib.parse.urlencode(parameters))
    print(url)
    return url

def response_func():
    global TOKEN
    token_url = TOKEN_URL.format(parameters['code'])
    payload = dict(PAYLOAD)
    payload['Accept'] = 'application/json'
    resp = s.get(TOKEN_URL.format(identifier), headers=payload)
    response = resp.json()
    TOKEN = response['authToken']
    return TOKEN

def return_token():
    token = None
    break_loop = False
    timeout = time.time() + 60*2
    while not break_loop:
        time.sleep(3)
        token = response_func()
        if token or time.time() > timeout:
            break_loop = True

    print(token)
    return token


PLEX_API_BASE_URL='https://plex.tv/users/account/?X-Plex-Token='
TOKEN = None


def check_plex_user():
    global TOKEN
    xml = requests.get(PLEX_API_BASE_URL + TOKEN)
    xml_parsed = ET.fromstring(xml.content)
    email = xml_parsed.attrib['email']
    id = xml_parsed.attrib['id']
    uuid = xml_parsed.attrib['uuid']
    username = xml_parsed.attrib['username']
    return { 'email': email, 'id': id, 'uuid': uuid, 'username': username }


