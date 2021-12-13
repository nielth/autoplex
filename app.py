from flask import Flask, jsonify, request
import rarbgapi

app = Flask(__name__)


@app.route('/<string:name>/', methods=['GET'])
def hello_word(name):
    client = rarbgapi.RarbgAPI()
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

    return torrent


if __name__ == '__main__':
    app.run(debug=True)
