import requests
import xmltodict

from flask import Flask
from flask import jsonify
from flask import request
from flask import abort

from flask_cors import CORS

from flask_limiter import Limiter
from flask_limiter.util import get_remote_address
from flask_jwt_extended import create_access_token
from flask_jwt_extended import set_access_cookies
from flask_jwt_extended import get_jwt_identity
from flask_jwt_extended import jwt_required
from flask_jwt_extended import JWTManager

plex_auth = "https://plex.tv/users/account/?X-Plex-Token="

app = Flask(__name__)

CORS(app, supports_credentials=True)
limiter = Limiter(
    get_remote_address,
    app=app,
    default_limits=["200 per day", "50 per hour"],
    storage_uri="memory://",
)


# Setup the Flask-JWT-Extended extension
app.config["JWT_TOKEN_LOCATION"] = [
    "headers", "cookies", "json", "query_string"]
# If true this will only allow the cookies that contain your JWTs to be sent
# over https. In production, this should always be set to True
app.config["JWT_COOKIE_SECURE"] = False

app.config["JWT_SECRET_KEY"] = "super-secret"  # Change this!

jwt = JWTManager(app)


# Create a route to authenticate your users and return JWTs. The
# create_access_token() function is used to actually generate the JWT.
@app.route("/login", methods=["POST"])
def login():
    username = request.json.get("username", None)
    password = request.json.get("password", None)
    if username != "test" or password != "test":
        return jsonify({"msg": "Bad username or password"}), 401

    access_token = create_access_token(identity=username)
    return jsonify(access_token=access_token)


# Protect a route with jwt_required, which will kick out requests
# without a valid JWT present.
@app.route("/protected", methods=["GET"])
@jwt_required()
def protected():
    # Access the identity of the current user with get_jwt_identity
    current_user = get_jwt_identity()
    return jsonify(logged_in_as=current_user), 200


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
