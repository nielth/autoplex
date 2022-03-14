import rarbgapi
import time
import os

from flask import Flask, request, render_template, redirect, url_for
from flask_sqlalchemy import SQLAlchemy
from dotenv import load_dotenv
from imdb import IMDb
from flask_login import (
    UserMixin,
    login_user,
    LoginManager,
    login_required,
    current_user,
    logout_user,
)

import _thread
import backend.torrent as torrent
import backend.logged_serv as logged_serv
import backend.plex_check as plex_check

load_dotenv()

app = Flask(__name__)
app.config["ENV"] = "development"
# python -c 'import secrets; print(secrets.token_hex())'
app.config["SECRET_KEY"] = os.getenv("SECRET_FLASK_KEY")
app.config["SESSION_COOKIE_SECURE"] = False
app.config["SESSION_COOKIE_HTTPONLY"] = True

client = rarbgapi.RarbgAPI()
ia = IMDb()
series = "tv series"
movie = "movie"

ADDRESS = os.getenv("ADDRESS")
PORT = os.getenv("APP_PORT")

login_manager = LoginManager()
login_manager.init_app(app)
login_manager.session_protection = "strong"


@login_manager.user_loader
def load_user(id):
    return User.query.get(int(id))


app.config["SQLALCHEMY_DATABASE_URI"] = "sqlite:///user.db"
app.config["SQLALCHEMY_TRACK_MODIFICATIONS"] = False
app.url_map.strict_slashes = False

db = SQLAlchemy(app)


class User(UserMixin, db.Model):
    id = db.Column(db.Integer, primary_key=True)
    email = db.Column(db.String(100), nullable=False, unique=True)
    uuid = db.Column(db.String(100), nullable=False, unique=True)
    username = db.Column(db.String(100), nullable=False, unique=True)


db.create_all()

FORWARD_LINK = None
PAYLOAD = {
    "X-Plex-Product": "Test Product",
    "X-Plex-Version": "0.0.1",
    "X-Plex-Device": "Test Device",
    "X-Plex-Platform": "Test Platform",
    "X-Plex-Device-Name": "Test Device Name",
    "X-Plex-Device-Vendor": "Test Vendor",
    "X-Plex-Model": "Test Model",
    "X-Plex-Client-Platform": "Test Client Platform",
}


@app.route("/", methods=["GET"])
@login_required
def index():
    return render_template("index.html")


@app.route("/logout")
def logout():
    logout_user()
    return redirect(url_for("login"))


@app.errorhandler(401)
def page_not_found(e):
    return redirect(url_for("login"))


@app.route("/login", methods=["GET"])
def login():
    FORWARD_LINK = plex_check.get_plex_link(
        forward_url=f"http://{ADDRESS}:{PORT}/login/callback"
    )
    return f'<a class="btn btn-success" href="{FORWARD_LINK}" target="_blank">Continue with Plex</a>'


@app.route("/login/callback", methods=["GET"])
def callback():
    token = plex_check.return_token()
    plex_user_info = plex_check.check_plex_user()
    plex_server_users = plex_check.get_server_accounts()
    username = plex_user_info["username"] in plex_server_users
    email = plex_user_info["email"] in plex_server_users
    if not any((username, email)):
        return "User is not linked to Plex Server", 404

    user = User.query.filter_by(email=plex_user_info["email"]).first()
    if user:
        login_user(user)
        return redirect(url_for("index"))
    else:
        new_user = User(
            email=plex_user_info["email"],
            id=plex_user_info["id"],
            uuid=plex_user_info["uuid"],
            username=plex_user_info["username"],
        )
        db.session.add(new_user)
        db.session.commit()
        user = User.query.filter_by(email=plex_user_info["email"]).first()
        login_user(user)
        return redirect(url_for("index"))


@app.route("/search/<string:category>/<string:title_search>", methods=["GET"])
@login_required
def search(category, title_search):
    return_category = category
    if category == "series":
        category = "tv " + category
    movie = ia.search_movie(title_search)
    info_list = dict()
    titles = 0

    for i in range(len(movie)):
        if (
            movie[i].data["kind"] == category
            and "Podcast" not in movie[i].data["title"]
            and "podcast" not in movie[i].data["title"]
            and "._V1_" in movie[i].data["cover url"]
        ):
            result = movie[i].data["cover url"].find("._V1_")
            info_list.setdefault("title", []).append(movie[i].data["title"])
            info_list.setdefault("cover", []).append(
                movie[i]
                .data["cover url"]
                .replace(movie[i].data["cover url"][result + 3 : result + 23], "")
            )
            info_list.setdefault("id", []).append(movie[i].movieID)
            titles += 1
    return render_template(
        "index.html",
        len=len(info_list["title"]),
        title=info_list["title"],
        cover=info_list["cover"],
        id=info_list["id"],
        category=return_category,
    )


def get_magnet(name, category):
    if category == "movie":
        return client.search(
            search_imdb=name,
            extended_response=True,
            sort="seeders",
            limit="100",
            categories=[
                rarbgapi.RarbgAPI.CATEGORY_MOVIE_H264_1080P,
                rarbgapi.RarbgAPI.CATEGORY_MOVIE_BD_REMUX,
            ],
        )
    elif category == "series":
        return client.search(
            search_imdb=name,
            extended_response=True,
            sort="seeders",
            limit="100",
            categories=[
                rarbgapi.RarbgAPI.CATEGORY_TV_EPISODES_HD,
                rarbgapi.RarbgAPI.CATEGORY_TV_EPISODES_UHD,
            ],
        )


@app.route("/torrent/<string:category>/<string:name>", methods=["GET"])
@login_required
def get_torrents(name, category):
    info_list = dict()
    title = ia.get_movie(name)
    result = title.data["cover url"].find("._V1_")
    cover = title.data["cover url"].replace(
        title.data["cover url"][result + 3 : result + 23], ""
    )
    torrent_search = "tt" + name
    titles = get_magnet(torrent_search, category)
    tries = 0

    while True:
        if not titles and tries <= 5:
            time.sleep(2)
            titles = get_magnet(torrent_search, category)
            tries += 1
        elif not titles and tries >= 5:
            return "", 404
        else:
            tries = 0
            break

    with open("backend/log/magnets.md", "a") as f:
        for i in range(len(titles)):
            if titles[i].seeders != 0:
                info_list.setdefault("filename", []).append(titles[i].filename)
                info_list.setdefault("size", []).append(
                    round((titles[i].size / 1073741824), 2)
                )
                info_list.setdefault("seeders", []).append(titles[i].seeders)
                info_list.setdefault("magnet", []).append(titles[i].download)
                f.write(titles[i].download + "\n")
                tries += 1

    return render_template(
        "index.html",
        len=len(info_list["magnet"]),
        filename=info_list["filename"],
        size=info_list["size"],
        seeders=info_list["seeders"],
        magnet=info_list["magnet"],
        cover=cover,
    )


@app.route("/torrent/<string:category>/<string:id>", methods=["POST"])
@login_required
def mov_magnet(category, id):
    magnet_link = request.form
    for line in reversed(open("backend/log/magnets.md").readlines()):
        if line.rstrip() in magnet_link["magnet"]:
            break
        else:
            return "", 404
    logged_serv.download_log(magnet_link, category)
    torrent.torrent_api(magnet_link["magnet"], category)
    return "", 204


if __name__ == "__main__":
    # _thread.start_new_thread(torrent.delete_finished, ())
    app.run(host="0.0.0.0", debug=True, threaded=True)
