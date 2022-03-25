import os

from flask import Flask, request, render_template, redirect, url_for, session
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

app = Flask(__name__)
app.config["ENV"] = "development"
# python -c 'import secrets; print(secrets.token_hex())'
app.config["SECRET_KEY"] = SECERT_KEY
app.config["SESSION_COOKIE_SECURE"] = True
app.config["SESSION_COOKIE_HTTPONLY"] = True
app.config["SQLALCHEMY_DATABASE_URI"] = "sqlite:///user.db"
app.config["SQLALCHEMY_TRACK_MODIFICATIONS"] = False
app.url_map.strict_slashes = False
# redirects stay https
app.wsgi_app = ProxyFix(app.wsgi_app, x_for=1 ,x_proto=1)

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
    forward_link, identifier = plex.get_plex_link(
        forward_url=f"{SECURITY}://{DOMAIN}/login/callback"
    )
    session["identifier"] = identifier
    return f'<a class="btn btn-success" href="{forward_link}" target="_blank">Continue with Plex</a>'


@app.route("/login/callback", methods=["GET"])
def callback():
    token = plex.return_token(session["identifier"])
    plex_user_info = plex.check_plex_user(token)
    try:
        plex_server_users = plex.get_server_accounts()
    except exceptions.ConnectionError:
        return "Plex server down, meaning I can't valid we're friends and you can't watch Plex :-(", 404
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
    t = titles.Titles
    info_list, return_category = t.get_titles(category, title_search)
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
    return render_template(
        "index.html",
        len=len(torrents["magnet"]),
        filename=torrents["filename"],
        size=torrents["size"],
        seeders=torrents["seeders"],
        magnet=torrents["magnet"],
        cover=cover,
        cat=category,
        na=name
    )


@app.route("/torrent/<string:category>/<string:name>", methods=["POST"])
@login_required
def mov_magnet(category, name):
    req_form = request.form
    magnet, title = req_form['magnet'], req_form['title']
    script_dir = os.path.dirname(__file__)
    abs_path = os.path.join(script_dir, "backend/log/magnets.txt")
    legit_magnet = False
    for line in reversed(open(abs_path).readlines()):
        temp = line.rstrip()
        if line.rstrip() in magnet:
            legit_magnet = True
            break
    if not legit_magnet:
        return "", 404
    #log.download_log(title, category)
    qbt.torrent_api(magnet, category)
    return "", 204


if __name__ == "__main__":
    p = Process(target = qbt.delete_finished)
    #p.start()
    app.run(host="0.0.0.0", debug=True)
