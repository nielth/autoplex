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

def main():
    global s
    s = requests.Session()
    resp = s.post(CODES_URL, data=PAYLOAD, headers=None)
    global identifier
    response = resp.json()
    parameters['code'] = response['code']
    identifier = response['id']
    url = AUTH_URL.format(urllib.parse.urlencode(parameters))
    print(url)
    temp_func()

def response_func():
    token_url = TOKEN_URL.format(parameters['code'])
    payload = dict(PAYLOAD)
    payload['Accept'] = 'application/json'
    resp = s.get(TOKEN_URL.format(identifier), headers=payload)
    response = resp.json()
    token = response['authToken']
    return token

def temp_func():
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

if __name__ == '__main__':
    main()
