import rarbgapi
import time

from flask import Flask, request, render_template
from imdb import IMDb

import _thread
import backend.torrent as torrent
import backend.logged_serv as logged_serv

app = Flask(__name__)
client = rarbgapi.RarbgAPI()
ia = IMDb()
series = "tv series"
movie = "movie"


@app.route("/", methods=["GET"])
def index():
    return render_template("index.html")


@app.route("/search/<string:category>/<string:title_search>", methods=["GET"])
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

    f = open("magnets.md", "a")
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
def mov_magnet(category, id):
    magnet_link = request.form
    for line in reversed(open("magnets.md").readlines()):
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
