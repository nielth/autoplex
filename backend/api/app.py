import requests
import xmltodict

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


@app.route("/hello", methods=["GET"])
def hello():
    current_user = get_jwt_identity()
    # check if user is loged in
    if current_user:
        print("True")
        return jsonify(logged_in_as=current_user), 200
    else:
        print("False")
        return ""


# Protect a route with jwt_required, which will kick out requests
# without a valid JWT present.
@app.route("/protected", methods=["GET", "POST"])
def protected():
    # Access the identity of the current user with get_jwt_identity
    try:
        verify_jwt_in_request()
        current_user = get_jwt_identity()
        return jsonify(logged_in_as=current_user), 200
    except:
        return ""


@app.route("/logout", methods=["POST"])
def logout_with_cookies():
    response = jsonify({"msg": "logout successful"})
    unset_jwt_cookies(response)
    return response


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


@app.route("/authToken", methods=["POST"])
def temp():
    data = request.json
    if not "authToken" in data:
        return ""

    user_req = req(plex_auth + data['authToken'])

    if user_req.status_code == 200:
        user_dump = xmltodict.parse(user_req.content)
    else:
        return ""
    if user_dump['user']['@id'] and user_dump['user']['@username']:
        user = user_dump['user']['@username']
        uid = user_dump['user']['@id']
    elif user_dump['user']['@id']:
        uid = user_dump['user']['@id']
    else:
        return ""

    server_req = req(
        "http://193.29.107.103:38983/accounts/?X-Plex-Token=nWWbRwG4matE1jXjBimL")
    if server_req.status_code == 200:
        server_dump = xmltodict.parse(server_req.content)
    else:
        return ""

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
