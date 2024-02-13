import requests
import json
import os
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

JWT_SECRET = os.getenv("JWT_SECRET")
OAUTH_FORWARD_URL = os.getenv("OAUTH_FORWARD_URL")

app.config["JWT_SECRET_KEY"] = JWT_SECRET
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


async def initiate_auth(
    forward_url: str,
    client_identifier=str(uuid4()),
):
    payload = PAYLOAD
    payload["X-Plex-Client-Identifier"] = client_identifier

    async with aiohttp.ClientSession().post(
        CODES_URL, data=payload, headers=None
    ) as resp:
        response = await resp.json()
        code = response["code"]
        identifier = response["id"]
    parameters = {
        "clientID": client_identifier,
        "code": code,
    }
    if forward_url:
        parameters["forwardUrl"] = forward_url

    url = AUTH_URL.format(urllib.parse.urlencode(parameters))
    return url, identifier, client_identifier


async def request_auth_token(identifier: str, client_identifier: str):
    """Request an auth token from Plex."""
    payload = dict(PAYLOAD)
    payload["X-Plex-Client-Identifier"] = client_identifier
    payload["Accept"] = "application/json"
    async with aiohttp.ClientSession().get(
        TOKEN_URL.format(identifier), headers=payload
    ) as resp:
        response = await resp.json()
        token = response["authToken"]
        return token


async def retrieve_token(identifier: str, client_identifier: str, timeout=60):
    """Poll Plex endpoint until a token is retrieved or times out."""
    token = None
    wait_until = datetime.now() + timedelta(seconds=timeout)
    break_loop = False
    while not break_loop:
        await sleep(3)
        token = await request_auth_token(identifier, client_identifier)
        if token or wait_until < datetime.now():
            break_loop = True

    return token


@app.route("/api/authToken", methods=["GET"])
async def authToken():
    data = await initiate_auth(
        forward_url=(
            OAUTH_FORWARD_URL
            if OAUTH_FORWARD_URL != ""
            else "https://autoplex.nielth.com/callback"
        )
    )
    response = make_response(jsonify({"url": data[0]}))
    response.set_cookie("identifier", str(data[1]))
    response.set_cookie("client_identifier", str(data[2]))
    return response


# 422 return code means too many requests
async def call(token: str):
    async with aiohttp.ClientSession() as session:
        async with session.get(ACCOUNT_URL.format(token)) as resp:
            resp_text = await resp.text()
            root = ET.fromstring(resp_text)
            username = root.get("username")
            email = root.get("email")
            uid = root.get("id")
        async with session.get(SERVER_URL) as server_resp:
            root = ET.fromstring(await server_resp.text())
            server_resp = root

    for account in server_resp.findall("Account"):
        server_name = account.get("name")
        server_id = account.get("id")
        if server_name in (username, email) or server_id == uid:
            return username

    return None


@app.route("/api/callback", methods=["GET"])
async def callback():
    identifier = request.cookies.get("identifier")
    client_identifier = request.cookies.get("client_identifier")
    token = await retrieve_token(
        identifier=identifier, client_identifier=client_identifier
    )
    username = await call(token=token)
    if not username:
        return "", 403
    response = jsonify(logged_in_as=username)
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
    return jsonify(logged_in_as=current_user), 200


@app.route("/api/logout", methods=["GET"])
@jwt_required()
def logout_with_cookies():
    verify_jwt_in_request()
    response = jsonify({"msg": "logout successful"})
    unset_jwt_cookies(response)
    return response


@app.route("/api/search/<string:search>/<string:page>", methods=["GET"])
@jwt_required()
def search_torrent(search: str, page: str = 0):
    session = requests.Session()

    # Load cookies from file
    cookies_path = Path("cookies.json")
    if cookies_path.exists():
        cookies = json.loads(cookies_path.read_text())
        cookies = requests.utils.cookiejar_from_dict(cookies)
        session.cookies.update(cookies)

    # Perform the GET request
    torrent_data = session.get(
        f"https://www.torrentleech.org/torrents/browse/list/categories/37,43,14,12,47,15,29,26,32,27/query/{search}/orderby/seeders/order/desc/page/{page}"
    )

    # Check and update cookies if new ones are received
    new_cookies = session.cookies.get_dict()
    if new_cookies:
        cookies_path.write_text(json.dumps(new_cookies, indent=4))

    # Return response
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
    response = session.get(torrent_link)
    if response.headers["Content-Type"] == "application/x-bittorrent":
        qbtTorrentDownload(response.content)
        return jsonify({"msg": True})

    return jsonify({"msg": False})
