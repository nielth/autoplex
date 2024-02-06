import requests
import json
import aiohttp
import urllib.parse
import xml.etree.ElementTree as ET

from asyncio import sleep
from datetime import datetime, timedelta
from uuid import uuid4
from pathlib import Path
from flask import Flask, jsonify, request, make_response
from datetime import datetime, timedelta
from flask_cors import CORS
from zoneinfo import ZoneInfo
from flask_jwt_extended import (
    get_jwt,
    create_access_token,
    set_access_cookies,
    get_jwt_identity,
    jwt_required,
    JWTManager,
    unset_jwt_cookies,
    verify_jwt_in_request,
)

from qbt.qbt import qbtTorrentDownload

app = Flask(__name__)

CORS(app, supports_credentials=True)

app.config["JWT_SECRET_KEY"] = "ecc0a726-7c69-4873-bd95-ebf8f0357ad1"
jwt = JWTManager(app)

app.config["JWT_TOKEN_LOCATION"] = ["cookies"]
app.config["JWT_COOKIE_SECURE"] = False
app.config["JWT_ACCESS_TOKEN_EXPIRES"] = timedelta(hours=5)


CODES_URL = "https://plex.tv/api/v2/pins.json?strong=true"
AUTH_URL = "https://app.plex.tv/auth#!?{}"
TOKEN_URL = "https://plex.tv/api/v2/pins/{}"
ACCOUNT_URL = "https://plex.tv/users/account/?X-Plex-Token={}"
SERVER_URL = "http://nielth.com:32444/accounts/?X-Plex-Token=9-2mk1MYy6BRgcGoG6SS"

PAYLOAD = {
    "X-Plex-Product": "Plex Auth App (Autoplex)",
    "X-Plex-Version": "0.69.420",
    "X-Plex-Device": "Linux",
    "X-Plex-Platform": "Linux",
    "X-Plex-Device-Name": "Autoplex",
    "X-Plex-Device-Vendor": "Test",
    "X-Plex-Model": "",
    "X-Plex-Client-Platform": "",
}


class PlexAuth:
    def __init__(
        self,
        payload,
        session=None,
        headers=None,
        base_url: str = None,
        identifier: str = None,
        client_identifier=str(uuid4()),
    ):
        """Create PlexAuth instance."""
        self.client_identifier = client_identifier
        self._code = None
        self._headers = headers
        self._identifier = identifier
        self._payload = payload
        self._base_url = base_url
        self._payload["X-Plex-Client-Identifier"] = self.client_identifier

        self._local_session = False
        self._session = session
        if session is None:
            self._session = aiohttp.ClientSession()
            self._local_session = True

    async def initiate_auth(self):
        """Request codes needed to create an auth URL. Starts external timeout."""
        async with self._session.post(
            CODES_URL, data=self._payload, headers=self._headers
        ) as resp:
            response = await resp.json()
            self._code = response["code"]
            self._identifier = response["id"]

    def auth_url(self, forward_url=None):
        """Return an auth URL for the user to follow."""
        parameters = {
            "clientID": self.client_identifier,
            "code": self._code,
            "forwardUrl": "http://localhost:8081/callback",
        }
        if forward_url:
            parameters["forwardUrl"] = forward_url

        url = AUTH_URL.format(urllib.parse.urlencode(parameters))
        return url, self._identifier, self.client_identifier

    async def request_auth_token(self):
        """Request an auth token from Plex."""
        payload = dict(self._payload)
        payload["Accept"] = "application/json"
        async with self._session.get(
            TOKEN_URL.format(self._identifier), headers=payload
        ) as resp:
            response = await resp.json()
            token = response["authToken"]
            return token

    async def token(self, timeout=60):
        """Poll Plex endpoint until a token is retrieved or times out."""
        token = None
        wait_until = datetime.now() + timedelta(seconds=timeout)
        break_loop = False
        while not break_loop:
            await sleep(3)
            token = await self.request_auth_token()
            if token or wait_until < datetime.now():
                break_loop = True

        return token

    async def close(self):
        """Close open client session."""
        if self._local_session:
            await self._session.close()

    async def __aenter__(self):
        """Async enter."""
        return self

    async def __aexit__(self, *exc_info):
        """Async exit."""
        await self.close()


async def test(base_url):
    async with PlexAuth(payload=PAYLOAD, base_url=base_url) as plexauth:
        await plexauth.initiate_auth()
        return plexauth.auth_url()


@app.route("/api/authToken", methods=["GET"])
async def authToken():
    data = await test(request.url)
    response = make_response(jsonify({"url": data[0]}))
    response.set_cookie("identifier", str(data[1]))
    response.set_cookie("client_identifier", str(data[2]))
    return response


async def test_2(identifier: str, client_identifier: str):
    async with PlexAuth(
        payload=PAYLOAD, identifier=identifier, client_identifier=client_identifier
    ) as plexauth:
        token = await plexauth.token()
        return token


# 422 return code means too many requests
async def call(token: str):
    async with aiohttp.ClientSession() as session:
        async with session.get(ACCOUNT_URL.format(token)) as resp:
            print(resp.status)
            resp_text = await resp.text()
            root = ET.fromstring(resp_text)
            username = root.get("username")
            email = root.get("email")
            uid = root.get("id")
        async with session.get(SERVER_URL) as server_resp:
            print(resp.status)
            root = ET.fromstring(await server_resp.text())
            server_resp = root

    for account in server_resp.findall("Account"):
        server_name = account.get("name")
        server_id = account.get("id")
        if server_name in (username, email) or server_id == uid:
            return True, username

    return False


@app.route("/api/callback", methods=["GET"])
async def callback():
    identifier = request.cookies.get("identifier")
    client_identifier = request.cookies.get("client_identifier")
    token = await test_2(identifier=identifier, client_identifier=client_identifier)
    test, username = await call(token=token)
    response = jsonify(logged_in_as=username, logged_in=test)
    access_token = create_access_token(identity=username)
    set_access_cookies(response, access_token)
    return response


def req(url):
    try:
        r = requests.get(url, timeout=1, verify=True)
        r.raise_for_status()
    except requests.exceptions.HTTPError as errh:
        print("HTTP Error")
        print(errh.args[0])
    except requests.exceptions.ReadTimeout as errrt:
        print("Time out")
    except requests.exceptions.ConnectionError as conerr:
        print("Connection error")
    except requests.exceptions.RequestException as errex:
        print("Exception request")
    return r


# Using an `after_request` callback, we refresh any token that is within 60 min
# minutes of expiring. Change the timedeltas to match the needs of application.
@app.after_request
def refresh_expiring_jwts(response):
    try:
        exp_timestamp = get_jwt()["exp"]
        now = datetime.now(tz=ZoneInfo("Europe/Oslo"))
        # print(exp_timestamp)
        target_timestamp = datetime.timestamp(now + timedelta(minutes=60))
        # target_timestamp = datetime.timestamp(now)
        if target_timestamp > exp_timestamp:
            access_token = create_access_token(identity=get_jwt_identity())
            set_access_cookies(response, access_token)
        return response
    except (RuntimeError, KeyError):
        # Case where there is not a valid JWT. Just return the original response
        return response


# Protect a route with jwt_required, which will kick out requests
# without a valid JWT present.
@app.route("/api/protected", methods=["GET"])
@jwt_required()
def protected():
    verify_jwt_in_request()
    current_user = get_jwt_identity()
    print(current_user)
    return jsonify(logged_in_as=current_user), 200


@app.route("/api/logout", methods=["POST"])
@jwt_required()
def logout_with_cookies():
    response = jsonify({"msg": "logout successful"})
    unset_jwt_cookies(response)
    return response


@app.route("/api/search", methods=["POST"])
@jwt_required()
def search_torrent():
    data = request.json
    session = requests.session()
    cookies = json.loads(Path("cookies.json").read_text())
    cookies = requests.utils.cookiejar_from_dict(cookies)
    session.cookies.update(cookies)
    torrent_data = session.get(
        f"https://www.torrentleech.org/torrents/browse/list/categories/37,43,14,12,47,15,29,26,32,27/query/{data['search']}/orderby/seeders/order/desc"
    )
    response = jsonify(torrent_data.json())
    return response


@app.route("/api/download", methods=["POST"])
@jwt_required()
def retrieve_torrent():
    data = request.json
    torrent_link = (
        f"https://www.torrentleech.org/download/{data['fid']}/{data['filename']}"
    )
    session = requests.session()
    cookies = json.loads(Path("cookies.json").read_text())
    cookies = requests.utils.cookiejar_from_dict(cookies)
    session.cookies.update(cookies)
    response = session.get(torrent_link, stream=True)
    qbtTorrentDownload(response.raw)
    response = jsonify({"msg": "torrent received"})
    return response


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000, debug=True)
