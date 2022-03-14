import requests
import xml.etree.ElementTree as ET
import time
import requests
import urllib.parse
import uuid
import os

from dotenv import load_dotenv


CODES_URL = 'https://plex.tv/api/v2/pins.json?strong=true'
AUTH_URL = 'https://app.plex.tv/auth#!?{}'
TOKEN_URL = 'https://plex.tv/api/v2/pins/{}'

PAYLOAD = {
    'X-Plex-Product': 'Plex Auth App (Autoplex)',
    'X-Plex-Version': '0.69.420',
    'X-Plex-Device': 'Linux',
    'X-Plex-Platform': 'Linux',
    'X-Plex-Device-Name': 'Autoplex',
    'X-Plex-Device-Vendor': '',
    'X-Plex-Model': '',
    'X-Plex-Client-Platform': '',
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


def get_server_accounts():
    headers = {"Accept": "application/json"}
    server_token = os.getenv('SERVER_TOKEN')
    accounts_link = f"http://10.0.0.33:32400/accounts/?X-Plex-Token={server_token}"
    resp = requests.get(accounts_link, headers=headers)
    response = resp.json()
    users = list()
    for user in response['MediaContainer']['Account']:
        users.append(user['name'])
    return users
    

