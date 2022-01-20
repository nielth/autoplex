from flask import Flask, request, render_template
import rarbgapi
from imdb import IMDb
from flask_cors import CORS
import time
import _thread

import torrent

app = Flask(__name__, static_folder="web", template_folder="web")
CORS(app)
client = rarbgapi.RarbgAPI()
ia = IMDb()
series = "tv series"
movie = "movie"


@app.route('/', methods=['GET'])
def index():
    return render_template("index.html")


@app.route('/search/<string:category>/<string:temp>/', methods=['GET'])
def search(category, temp):
    return_category = category
    if(category == "series"):
        category = "tv " + category
    movie = ia.search_movie(temp)
    title = list()
    cover = list()
    id = list()
    titles = 0

    for i in range(len(movie)):
        if movie[i].data['kind'] == category and "Podcast" not in movie[i].data['title'] and "podcast" not in \
                movie[i].data['title'] and "._V1_" in movie[i].data['cover url']:
            result = movie[i].data['cover url'].find("._V1_")
            title.append(movie[i].data['title'])
            cover.append(movie[i].data['cover url'].replace(
                movie[i].data['cover url'][result + 3:result + 23], ""))
            id.append(movie[i].movieID)
            titles += 1
    return render_template('index.html', len=len(title), title=title, cover=cover, id=id, category=return_category)


def get_magnet(name, category):
    if category == "movie":
        return client.search(search_imdb=name, extended_response=True, sort="seeders", limit="100",
                             categories=[rarbgapi.RarbgAPI.CATEGORY_MOVIE_H264_1080P,
                                         rarbgapi.RarbgAPI.CATEGORY_MOVIE_BD_REMUX])
    elif category == "series":
        return client.search(search_imdb=name, extended_response=True, sort="seeders", limit="100",
                             categories=[rarbgapi.RarbgAPI.CATEGORY_TV_EPISODES_HD,
                                         rarbgapi.RarbgAPI.CATEGORY_TV_EPISODES_UHD])


@app.route('/torrent/<string:category>/<string:name>/', methods=['GET'])
def get_torrents(name, category):
    filename = list()
    size = list()
    seeders = list()
    magnet = list()
    title = ia.get_movie(name)
    result = title.data['cover url'].find("._V1_")
    cover = title.data['cover url'].replace(
        title.data['cover url'][result + 3:result + 23], "")
    torrent_search = "tt" + name
    titles = get_magnet(torrent_search, category)
    tries = 0

    while True:
        if not titles and tries <= 5:
            time.sleep(2)
            titles = get_magnet(torrent_search, category)
            tries += 1
        else:
            tries = 0
            break

    f = open("magnets.md", "a")
    for i in range(len(titles)):
        if titles[i].seeders != 0:
            filename.append(titles[i].filename)
            size.append(round((titles[i].size / 1073741824), 2))
            seeders.append(titles[i].seeders)
            magnet.append(titles[i].download)
            f.write(titles[i].download + "\n")
            tries += 1
    f.close()

    return render_template('index.html', len=len(magnet), filename=filename, size=size, seeders=seeders, magnet=magnet, cover=cover)


@app.route('/torrent/<string:category>/<string:id>/', methods=['POST'])
def mov_magnet(category, id):
    tries = 0
    magnet_link = request.form
    for line in reversed(open("magnets.md").readlines()):
        if line.rstrip() in magnet_link['magnet']:
            break
        elif tries == 100:
            return "", 404
        tries += 1
    torrent.torrentAPI(magnet_link['magnet'], category)
    return "", 204


if __name__ == '__main__':
    _thread.start_new_thread(torrent.deleteFinished, ())
    app.run(host="0.0.0.0")
