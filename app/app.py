import rarbgapi
import time
import flask
import os


from flask_sqlalchemy import SQLAlchemy
from dotenv import load_dotenv
from flask_wtf import FlaskForm
from wtforms import PasswordField
from wtforms.validators import DataRequired, Email, Length
from wtforms.fields import EmailField
from werkzeug.security import generate_password_hash, check_password_hash
from flask_login import (
    UserMixin,
    login_user,
    LoginManager,
    login_required,
    current_user,
    logout_user
)
from flask import Flask, request, render_template, redirect, url_for
from imdb import IMDb

import _thread
import backend.torrent as torrent
import backend.logged_serv as logged_serv
import backend.plex_check as plex_check

import asyncio
from plexauth import PlexAuth

load_dotenv()

app = Flask(__name__)
app.config['ENV'] = 'development'
# python -c 'import secrets; print(secrets.token_hex())'
app.config['SECRET_KEY'] = os.getenv('SECRET_FLASK_KEY')
app.config['SESSION_COOKIE_SECURE'] = True
app.config['SESSION_COOKIE_HTTPONLY'] = True

client = rarbgapi.RarbgAPI()
ia = IMDb()
series = "tv series"
movie = "movie"

login_manager = LoginManager()
login_manager.init_app(app)
login_manager.session_protection = "strong"


@login_manager.user_loader
def load_user(id):
    return User.query.get(int(id))


app.config['SQLALCHEMY_DATABASE_URI'] = 'sqlite:///user.db'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
app.url_map.strict_slashes = False

db = SQLAlchemy(app)


class User(UserMixin, db.Model):
    id = db.Column(db.Integer, primary_key=True)
    email = db.Column(db.String(100), nullable=False, unique=True)

db.create_all()

PAYLOAD = {
    'X-Plex-Product': 'Test Product',
    'X-Plex-Version': '0.0.1',
    'X-Plex-Device': 'Test Device',
    'X-Plex-Platform': 'Test Platform',
    'X-Plex-Device-Name': 'Test Device Name',
    'X-Plex-Device-Vendor': 'Test Vendor',
    'X-Plex-Model': 'Test Model',
    'X-Plex-Client-Platform': 'Test Client Platform'
}

@app.route("/", methods=["GET"])
@login_required
def index():
    return render_template("index.html")


@app.errorhandler(401)
def page_not_found(e):
    return redirect(url_for('login'))


loop = asyncio.get_event_loop()


@app.route("/login", methods=['GET'])
async def login():
    plexauth = PlexAuth(PAYLOAD)
    await plexauth.initiate_auth()
    print(plexauth.auth_url(forward_url="http://localhost:1337"))
    return f'<a href"{plexauth.auth_url(forward_url="http://localhost:1337")} target="_blank">Link</a>"'
    token = await plexauth.token()
    if token:
        print("Token: {}".format(token))
    else:
        print("No token returned.")
    time.sleep(30)
    user = User.query.filter_by(email="thomas.nielsen98@gmail.com").first()
    if 0:
        if 1:
            login_user(user)
            return redirect(url_for('index'))
    else:
        flask.flash("user not found")

    return render_template("index.html")

@app.route("/register", methods=['GET', 'POST'])
def register():
    form = RegisterForm()
    if form.validate_on_submit():
        email = form.email.data
        password = form.password.data
        if User.query.filter_by(email=email).first():
            flask.flash("user has already registered")
        else:
            # create new user
            new_user = User(
                email = email,
                password = generate_password_hash(password, method='pbkdf2:sha256', salt_length=8)
            )
            db.session.add(new_user)
            db.session.commit()
            login_user(new_user)
            return redirect(url_for('index'))

    return render_template("register.html", form=form)


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
            info_list.setdefault("cover", []).append(movie[i]
                                                     .data["cover url"]
                                                     .replace(movie[i].data["cover url"][result + 3: result + 23], ""))
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
        title.data["cover url"][result + 3: result + 23], ""
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

    f = open("../log/magnets.md", "a")
    for i in range(len(titles)):
        if titles[i].seeders != 0:
            info_list.setdefault("filename", []).append(titles[i].filename)
            info_list.setdefault("size", []).append(
                round((titles[i].size / 1073741824), 2))
            info_list.setdefault("seeders", []).append(titles[i].seeders)
            info_list.setdefault("magnet", []).append(titles[i].download)
            f.write(titles[i].download + "\n")
            tries += 1
    f.close()

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
    for line in reversed(open("./log/magnets.md").readlines()):
        if line.rstrip() in magnet_link["magnet"]:
            break
        else:
            return "", 404
    logged_serv.download_log(magnet_link, category)
    torrent.torrent_api(magnet_link["magnet"], category)
    return "", 204


if __name__ == "__main__":
    _thread.start_new_thread(torrent.delete_finished, ())
    app.run(host="0.0.0.0", debug=True)
