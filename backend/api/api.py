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

OAUTH_FORWARD_URL = os.getenv("OAUTH_FORWARD_URL")
PLEX_URL = os.getenv("PLEX_URL")
PLEX_TOKEN = os.getenv("PLEX_TOKEN")
TL_USER = os.getenv("TL_USER")
TL_PASS = os.getenv("TL_PASS")

app.config["JWT_SECRET_KEY"] = (
    os.urandom(16)
    if not os.getenv("FLASK_ENV") == "development"
    else "secret_key_for_dev"
)
jwt = JWTManager(app)

app.config["JWT_TOKEN_LOCATION"] = ["cookies"]

if os.getenv("FLASK_ENV") == "development":
    app.config["JWT_COOKIE_SECURE"] = False
else:
    app.config["JWT_COOKIE_SECURE"] = True


app.config["JWT_ACCESS_TOKEN_EXPIRES"] = timedelta(hours=24 * 14)

CODES_URL = "https://plex.tv/api/v2/pins.json?strong=true"
AUTH_URL = "https://app.plex.tv/auth#!?{}"
TOKEN_URL = "https://plex.tv/api/v2/pins/{}"
ACCOUNT_URL = "https://plex.tv/users/account/?X-Plex-Token={}"
SERVER_URL = f"{PLEX_URL}/accounts/?X-Plex-Token={PLEX_TOKEN}"

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


async def initiate_auth(
    forward_url: str,
    client_identifier=str(uuid4()),
):
    payload = PAYLOAD
    payload["X-Plex-Client-Identifier"] = client_identifier

    async with aiohttp.ClientSession() as session:
        async with session.post(CODES_URL, data=payload, headers=None) as resp:
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
    async with aiohttp.ClientSession() as session:
        async with session.get(TOKEN_URL.format(identifier), headers=payload) as resp:
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
    data = await initiate_auth(forward_url=f"{OAUTH_FORWARD_URL}/callback")
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


def tl_search(search, page):
    response = ""
    for _ in range(2):
        session = requests.Session()

        # Load cookies from file
        cookies_path = Path("cookies.json")
        if cookies_path.exists():
            cookies = json.loads(cookies_path.read_text())
            cookies = requests.utils.cookiejar_from_dict(cookies)
            session.cookies.update(cookies)
        

        ua = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"

        # Perform the GET request
        response = session.get(
            f"https://www.torrentleech.org/torrents/browse/list/categories/37,43,14,12,47,15,29,26,32,27,44,36/query/{search}/orderby/seeders/order/desc/page/{page}",
            headers={
                "User-Agent": ua
            },
        )

        print("Get:", response.status_code, response.request.headers)

        if response.status_code == 403:
            url = "http://flaresolverr:8191/v1"
            headers = {"Content-Type": "application/json"}
            data = {
                "cmd": "request.get",
                "url": "https://www.torrentleech.org/",
                "maxTimeout": 60000
            }
            response1 = requests.post(url, headers=headers, json=data)

            if response1.status_code == 200:
                resp_json = response1.json()
                cookies = resp_json["solution"]["cookies"]
                ua = resp_json["solution"]["userAgent"]
                print("login:", ua, cookies)
                for cookie in cookies:
                    session.cookies.set(cookie["name"], cookie["value"])

            login_data = {
                "username": TL_USER,
                "password": TL_PASS
            }

            login_url = 'https://www.torrentleech.org/user/account/login/'
            response2 = session.post(login_url, data=login_data, headers={"User-Agent": ua })

            print(response2.status_code)


        new_cookies = session.cookies.get_dict()
        if new_cookies:
            cookies_path.write_text(json.dumps(new_cookies, indent=4))

        if response.status_code == 200:
            break

    return response


@app.route("/api/search/<string:search>/<string:page>", methods=["GET"])
@jwt_required()
def search_torrent(search: str, page: str = "0"):
    torrent_data = tl_search(search, page)
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
    response = session.get(
        torrent_link,
        headers={
            "User-Agent": "Mozilla/5.0 (X11; Linux x86_64; rv:130.0) Gecko/20100101 Firefox/130.0"
        },
    )

    if data["categoryID"] in (8, 9, 11, 37, 43, 14, 12, 13, 47, 15, 29, 36):
        category = "movies"
    elif data["categoryID"] in (26, 32, 27, 44):
        category = "tvseries"

    if response.headers["Content-Type"] == "application/x-bittorrent":
        qbtTorrentDownload(response.content, category)
        return jsonify({"msg": True})

    return jsonify({"msg": False})


@app.route("/api/cookie", methods=["GET", "POST"])
def cookie():
    cookie = request.args.get("cookie")
    print(cookie, flush=True)
    return jsonify({"msg": False})
