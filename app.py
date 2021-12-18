from flask import Flask, jsonify, request, Response
import rarbgapi
from imdb import IMDb
import torrent_api

app = Flask(__name__)
client = rarbgapi.RarbgAPI()
ia = IMDb()


@app.route('/torrent/<string:name>/', methods=['GET'])
def hello_word(name):
    titles = client.search(search_imdb=name, extended_response=True, sort="seeders",
                           categories=[rarbgapi.RarbgAPI.CATEGORY_MOVIE_H264_1080P,
                                       rarbgapi.RarbgAPI.CATEGORY_MOVIE_BD_REMUX])

    torrent = dict()

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


@app.route('/imdb/<string:name>/', methods=['GET'])
def imdb_mov(name):
    name.replace("%20", " ")
    movie = ia.search_movie(name)
    movies = dict()
    mov_list = 0

    for i in range(len(movie)):
        if movie[i].data['kind'] == 'movie':
            result = movie[i].data['cover url'].find("._V1_")
            movies[mov_list] = ({"title": movie[i].data['title'],
                                 "cover-url": movie[i].data['cover url'].replace(
                                     movie[i].data['cover url'][result + 3:result + 23], ""),
                                 "movie-id": "tt" + movie[i].movieID,
                                 })
            mov_list += 1

    response = jsonify(movies)
    response.headers.add('Access-Control-Allow-Origin', '*')
    return response


@app.route('/movie/', methods=['POST'])
def mov_magnet():
    magnet_link = request.form
    torrent_api.add_movie_to_api(magnet_link['magnet'])
    return "", 204


if __name__ == '__main__':
    app.run(debug=True)
