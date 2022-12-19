import os
import json

from flask import Flask, request, render_template, redirect, url_for, session, jsonify
from werkzeug.security import generate_password_hash, check_password_hash
from werkzeug.middleware.proxy_fix import ProxyFix
from flask_sqlalchemy import SQLAlchemy
from requests import exceptions
from multiprocessing import Process
from dotenv import load_dotenv
from threading import Thread
from flask_login import (
    UserMixin,
    login_user,
    LoginManager,
    login_required,
    current_user,
    logout_user,
)

import backend.qbt as qbt
import backend.log as log
import backend.plex as plex
import backend.titles as titles

load_dotenv()

DOMAIN = os.getenv("DOMAIN")
SECERT_KEY = os.getenv("SECRET_FLASK_KEY")
SECURITY = os.getenv("SECURITY")
REGISTER_KEY = os.getenv("REGISTER_KEY")

app = Flask(__name__)
app.config["ENV"] = "development"
# python3 -c 'import secrets; print(secrets.token_hex())'
app.config["SECRET_KEY"] = SECERT_KEY
app.config["SESSION_COOKIE_SECURE"] = True
app.config["SESSION_COOKIE_HTTPONLY"] = True
app.config["SQLALCHEMY_DATABASE_URI"] = "sqlite:///user.db"
app.config["SQLALCHEMY_TRACK_MODIFICATIONS"] = False

# For debugging
# app.config["TEMPLATES_AUTO_RELOAD=True"] = True
# app.config["DEBUG"] = True

app.url_map.strict_slashes = False
# redirects stay https
app.wsgi_app = ProxyFix(app.wsgi_app, x_for=1, x_proto=1)

db = SQLAlchemy(app)
from backend.db_ import *

series = "tv series"
movie = "movie"

login_manager = LoginManager()
login_manager.init_app(app)
login_manager.session_protection = "strong"


@login_manager.user_loader
def load_user(id):
    return User.query.get(int(id))


@app.route("/", methods=["GET"])
@login_required
def index():
    if current_user.is_authenticated:
        return render_template("index.html")
    else:
        return render_template("login.html")


@app.route("/logout")
def logout():
    logout_user()
    return redirect(url_for("login"))


@app.errorhandler(401)
def page_not_found(e):
    return redirect(url_for("login"))


@app.errorhandler(404)
def not_found_error(error):
    return render_template("404.html"), 404


@app.errorhandler(500)
def not_found_error(error):
    return render_template("404.html"), 500


@app.route("/login", methods=["GET"])
def login():
    forward_link, identifier = plex.get_plex_link(
        forward_url=f"https://{DOMAIN}/login/callback"
    )
    session["identifier"] = identifier
    return render_template("connect.html", forward_link=forward_link)


@app.route("/login/nonauth", methods=["GET", "POST"])
def sign_in():
    if request.method == "GET":
        return render_template("login.html")
    else:
        email = request.form.get("email")
        password = request.form.get("password")
        user = User.query.filter_by(email=email).first()
        try:
            passwd = user.password
        except AttributeError:
            return (
                "Username or password wrong or not found. Maybe try logging in through Plex?",
                404,
            )
        if not user or not passwd:
            return "User not found", 404
        elif not check_password_hash(passwd, password):
            return (
                "Username or password wrong or not found. Maybe try logging in through Plex?",
                404,
            )
        login_user(user, remember=True)
        return redirect(url_for("index"))


@app.route(f"/signup/{REGISTER_KEY}", methods=["GET", "POST"])
def signup():
    global REGISTER_KEY
    if request.method == "GET":
        return render_template("signup.html", REGISTER_KEY=REGISTER_KEY)
    else:
        email = request.form.get("email")
        password = request.form.get("password")
        user = User.query.filter_by(email=email).first()
        if user:
            user.password = generate_password_hash(password, method="sha256")
            db.session.commit()
            login_user(user, remember=True)
            return redirect(url_for("index"))
        new_user = User(
            email=email, password=generate_password_hash(password, method="sha256")
        )
        db.session.add(new_user)
        db.session.commit()
        login_user(user, remember=True)
        return redirect(url_for("index"))


@app.route("/login/callback", methods=["GET"])
def callback():
    try:
        sess = session["identifier"]
    except KeyError:
        return redirect(url_for("index"))
    token = plex.return_token(sess)
    plex_user_info = plex.check_plex_user(token)
    try:
        plex_server_users, plex_server_id = plex.get_server_accounts()
    except exceptions.ConnectionError:
        return (
            "Cannot contact my Plex server :-( Alert yours truly if persistent.",
            404,
        )

    id_plex = plex_user_info["id"] in plex_server_id
    username = plex_user_info["username"] in plex_server_users
    email = plex_user_info["email"] in plex_server_users
    if not any((id_plex, username, email)):
        return "User is not linked to Plex Server", 404

    user = User.query.filter_by(email=plex_user_info["email"]).first()
    if user and user:
        login_user(user, remember=True)
        if user.username is None:
            user.id = plex_user_info["id"]
            user.uuid = plex_user_info["uuid"]
            user.username = plex_user_info["username"]
            db.session.commit()
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
        login_user(user, remember=True)
        return redirect(url_for("index"))


@app.route("/search/<string:category>/<string:title_search>", methods=["GET"])
@login_required
def search(category, title_search):
    t = titles.Titles
    info_list, return_category = t.get_titles(category, title_search)
    if info_list is False:
        return "Had problems connecting to IMDB to retrieve movies and series", 404
    return render_template(
        "index.html",
        len=len(info_list["title"]),
        title=info_list["title"],
        cover=info_list["cover"],
        id=info_list["id"],
        category=return_category,
    )


@app.route("/torrent/<string:category>/<string:name>", methods=["GET"])
@login_required
def get_torrents(name, category):
    t = titles.Titles
    torrents, cover = t.torrents(name, category)
    if torrents is False:
        return "Had problems connecting to IMDB to retrieve movies and series", 404
    return render_template(
        "index.html",
        len=len(torrents["magnet"]),
        filename=torrents["filename"],
        size=torrents["size"],
        seeders=torrents["seeders"],
        magnet=torrents["magnet"],
        cover=cover,
        cat=category,
        na=name,
    )


@app.route("/torrent/<string:category>/<string:name>", methods=["POST"])
@login_required
def mov_magnet(category, name):
    req_form = request.form
    magnet, title = req_form["magnet"], req_form["title"]
    log.store_magnet(magnet)
    log.log_user_download(current_user, category, title, magnet)
    qbt.torrent_api(magnet, category)
    return "", 204


@app.route("/downloading", methods=["GET"])
@login_required
def downloading():
    return_title, return_status = qbt.download_status()
    return render_template(
        "download.html",
        len=len(return_title),
        title=return_title,
        progress=return_status,
    )


@app.route("/log", methods=["GET"])
@login_required
def download_log():
    logged_download = log.get_user_download()
    return jsonify(logged_download)


if __name__ == "__main__":
    app.run(host="0.0.0.0")
