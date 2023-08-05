import requests
import xmltodict
import json
import shutil

from pathlib import Path

from flask import Flask
from flask import jsonify
from flask import request
from flask import abort

from datetime import datetime
from datetime import timedelta
from datetime import timezone

from flask_cors import CORS

from flask_limiter import Limiter
from flask_limiter.util import get_remote_address
from flask_jwt_extended import get_jwt
from flask_jwt_extended import create_access_token
from flask_jwt_extended import set_access_cookies
from flask_jwt_extended import get_jwt_identity
from flask_jwt_extended import jwt_required
from flask_jwt_extended import JWTManager
from flask_jwt_extended import unset_jwt_cookies
from flask_jwt_extended import verify_jwt_in_request

from qbt.qbt import qbtTorrentDownload

plex_auth = "https://plex.tv/users/account/?X-Plex-Token="

app = Flask(__name__)

CORS(app, supports_credentials=True)
limiter = Limiter(
    get_remote_address,
    app=app,
    #default_limits=["200 per day", "50 per hour"],
    storage_uri="memory://",
)


# Setup the Flask-JWT-Extended extension
app.config["JWT_TOKEN_LOCATION"] = ["cookies"]
# If true this will only allow the cookies that contain your JWTs to be sent
# over https. In production, this should always be set to True
app.config["JWT_COOKIE_SECURE"] = False
app.config["JWT_ACCESS_TOKEN_EXPIRES"] = timedelta(hours=1)

app.config["JWT_SECRET_KEY"] = "super-secret"  # Change this!

jwt = JWTManager(app)


def req(url):
    try:
        r = requests.get(
            url, timeout=1, verify=True)
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

# Using an `after_request` callback, we refresh any token that is within 30
# minutes of expiring. Change the timedeltas to match the needs of your application.


@app.after_request
def refresh_expiring_jwts(response):
    try:
        exp_timestamp = get_jwt()["exp"]
        now = datetime.now(timezone.utc)
        target_timestamp = datetime.timestamp(now + timedelta(minutes=30))
        if target_timestamp > exp_timestamp:
            access_token = create_access_token(identity=get_jwt_identity())
            set_access_cookies(response, access_token)
        return response
    except (RuntimeError, KeyError):
        # Case where there is not a valid JWT. Just return the original response
        return response


def get_identity_if_logedin():
    try:
        return get_jwt_identity()
    except Exception:
        pass


# Protect a route with jwt_required, which will kick out requests
# without a valid JWT present.
@app.route("/protected", methods=["GET", "POST"])
def protected():
    # Access the identity of the current user with get_jwt_identity
    try:
        verify_jwt_in_request()
        current_user = get_jwt_identity()
        print(current_user)
        return jsonify(logged_in_as=current_user), 200
    except:
        return ""


@app.route("/logout", methods=["POST"])
@jwt_required()
def logout_with_cookies():
    response = jsonify({"msg": "logout successful"})
    unset_jwt_cookies(response)
    return response


@app.route("/search", methods=["POST"])
@jwt_required()
def search_torrent():
    data = request.json
    session = requests.session()
    cookies = json.loads(Path("cookies.json").read_text())
    cookies = requests.utils.cookiejar_from_dict(cookies)
    session.cookies.update(cookies)
    torrent_data = session.get(
        f"https://www.torrentleech.org/torrents/browse/list/categories/37,43,14,12,47,15,29,26,32,27/query/{data['search']}/orderby/seeders/order/desc")
    response = jsonify(torrent_data.json())
    return response


@app.route("/download", methods=["POST"])
@jwt_required()
def retrieve_torrent():
    data = request.json
    torrent_link = f"https://www.torrentleech.org/download/{data['fid']}/{data['filename']}"
    session = requests.session()
    cookies = json.loads(Path("cookies.json").read_text())
    cookies = requests.utils.cookiejar_from_dict(cookies)
    session.cookies.update(cookies)
    response = session.get(torrent_link, stream=True)
    qbtTorrentDownload(response.raw)
    response = jsonify({"msg": "torrent received"})
    return response


@app.route("/authToken", methods=["POST"])
def authToken():
    data = request.json
    if not "authToken" in data:
        return "No authToken", 401

    user_req = req(plex_auth + data['authToken'])

    if user_req.status_code == 200:
        user_dump = xmltodict.parse(user_req.content)
    else:
        return f"could not connect to: {plex_auth}", 401
    if user_dump['user']['@id'] and user_dump['user']['@username']:
        user = user_dump['user']['@username']
        uid = user_dump['user']['@id']
    elif user_dump['user']['@id']:
        uid = user_dump['user']['@id']
    else:
        return "Could not find user on server", 401

    server_req = req(
        "http://178.232.41.76:32444/accounts/?X-Plex-Token=nWWbRwG4matE1jXjBimL")
    if server_req.status_code == 200:
        server_dump = xmltodict.parse(server_req.content)
    else:
        return "Could not connect to http://178.232.41.76:32444", 401

    for server_user in server_dump['MediaContainer']['Account']:
        if uid == server_user['@id'] or user == server_user['@name']:
            print("Success!")
            response = jsonify({"msg": "login successful"})
            access_token = create_access_token(identity=user)
            set_access_cookies(response, access_token)
            return response

    return jsonify({"msg": "User not connected to Plex server"}), 401


if __name__ == "__main__":
    app.run(host='0.0.0.0', debug=True)
