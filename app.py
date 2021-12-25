from flask import Flask, jsonify, request, render_template
import rarbgapi
from imdb import IMDb
from flask_cors import CORS
import time

import torrent_api

app = Flask(__name__, static_folder="web", template_folder="web")
CORS(app)
client = rarbgapi.RarbgAPI()
ia = IMDb()
series = "tv series"
movie = "movie"


@app.route('/')
def index():
    return render_template("index.html")


def get_title(name, category):
    name.replace("%20", " ")
    movie = ia.search_movie(name)
    movies = dict()
    titles = 0

    for i in range(len(movie)):
        if movie[i].data['kind'] == category and "Podcast" not in movie[i].data['title'] and "podcast" not in \
                movie[i].data['title']:
            result = movie[i].data['cover url'].find("._V1_")
            movies[titles] = ({"title": movie[i].data['title'],
                               "cover-url": movie[i].data['cover url'].replace(
                                   movie[i].data['cover url'][result + 3:result + 23], ""),
                               "movie-id": "tt" + movie[i].movieID,
                               })
            titles += 1

    response = jsonify(movies)
    response.headers.add('Access-Control-Allow-Origin', '*')
    return response


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
            break

    for i in range(len(titles)):
        torrent[i] = ({"filename": titles[i].filename,
                       "size": round((titles[i].size / 1073741824), 2),
                       "category": titles[i].category,
                       "seeders": titles[i].seeders,
                       "magnet": titles[i].download
                       })

    response = jsonify(torrent)
    response.headers.add('Access-Control-Allow-Origin', '*')
    return response


def get_magnet(name, category):
    if category == "movie":
        return client.search(search_imdb=name, extended_response=True, sort="seeders",
                             categories=[rarbgapi.RarbgAPI.CATEGORY_MOVIE_H264_1080P,
                                         rarbgapi.RarbgAPI.CATEGORY_MOVIE_BD_REMUX])
    elif category == "series":
        return client.search(search_imdb="tt0944947", extended_response=True, sort="seeders")


@app.route('/imdb/movie/<string:name>/', methods=['GET'])
def imdb_mov(name):
    return get_title(name, movie)


@app.route('/imdb/series/<string:name>/', methods=['GET'])
def imdb_series(name):
    return get_title(name, series)


@app.route('/<string:category>/', methods=['POST'])
def mov_magnet(category):
    magnet_link = request.form
    torrent_api.add_movie_to_api(magnet_link['magnet'], category)
    return "", 204


if __name__ == '__main__':
    app.run(host="0.0.0.0", debug=True)
