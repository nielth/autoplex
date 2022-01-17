from flask import Flask, jsonify, request, render_template
import rarbgapi
from imdb import IMDb
from flask_cors import CORS
import time

import torrent

app = Flask(__name__, static_folder="web", template_folder="web")
CORS(app)
client = rarbgapi.RarbgAPI()
ia = IMDb()
series = "tv series"
movie = "movie"


@app.route('/')
def index():
    return render_template("index.html")


@app.route('/search/')
def search():
    temp = request.args.get('search')
    is_movie = request.args.get('movie')
    is_series = request.args.get('movie')
    if is_movie == "on":
        get_title(temp, "movie")
    elif is_series == "on":
        get_title(temp, "series")


def get_title(temp, category):
    temp.replace("%20", " ")
    movie = ia.search_movie(temp)
    movies = dict()
    title = list()
    cover = list()
    id = list()
    titles = 0

    for i in range(len(movie)):
        if "Podcast" not in movie[i].data['title'] and "podcast" not in \
                movie[i].data['title']:
            result = movie[i].data['cover url'].find("._V1_")
            title.append(movie[i].data['title'])
            cover.append(movie[i].data['cover url'].replace(
                                   movie[i].data['cover url'][result + 3:result + 23], ""))
            id.append(movie[i].movieID)
            titles += 1

    response = jsonify(movies)
    return render_template('index.html', len = len(title), title=title, cover=cover, id=id)

@app.route('/torrent/<string:category>/<string:name>/', methods=['GET'])
def get_torrents(name, category):
    torrent = dict()
    titles = get_magnet(name, category)
    tries = 0

    while True:
        if not titles and tries <= 5:
            time.sleep(2)
            titles = get_magnet(name, category)
            tries += 1
        else:
            tries = 0
            break

    for i in range(len(titles)):
        if titles[i].seeders != 0:
            torrent[tries] = ({"filename": titles[i].filename,
                               "size": round((titles[i].size / 1073741824), 2),
                               "category": titles[i].category,
                               "seeders": titles[i].seeders,
                               "magnet": titles[i].download
                               })
            tries += 1

    response = jsonify(torrent)
    response.headers.add('Access-Control-Allow-Origin', '*')
    return response


def get_magnet(name, category):
    if category == "movie":
        return client.search(search_imdb=name, extended_response=True, sort="seeders", limit="100",
                             categories=[rarbgapi.RarbgAPI.CATEGORY_MOVIE_H264_1080P,
                                         rarbgapi.RarbgAPI.CATEGORY_MOVIE_BD_REMUX])
    elif category == "series":
        return client.search(search_imdb=name, extended_response=True, sort="seeders", limit="100",
                             categories=[rarbgapi.RarbgAPI.CATEGORY_TV_EPISODES_HD,
                                         rarbgapi.RarbgAPI.CATEGORY_TV_EPISODES_UHD])


@app.route('/imdb/movie/<string:name>/', methods=['GET'])
def imdb_mov(name):
    return get_title(name, movie)


@app.route('/imdb/series/<string:name>/', methods=['GET'])
def imdb_series(name):
    return get_title(name, series)


@app.route('/<string:category>/', methods=['POST'])
def mov_magnet(category):
    magnet_link = request.form
    torrent.torrentAPI(magnet_link['magnet'], category)
    return "", 204


if __name__ == '__main__':
    app.run(host="0.0.0.0", debug=True)
