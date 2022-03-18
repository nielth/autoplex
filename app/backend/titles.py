from imdb import IMDb
import rarbgapi
import time
import os

ia = IMDb()
client = rarbgapi.RarbgAPI()


class Titles:
    def __init__(self):
        pass


    def get_titles(category, title_search):
        info_list = dict()
        return_category = category
        if category == "series":
            category = "tv " + category
        movie = ia.search_movie(title_search)
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
        return info_list, return_category


    def torrents(name, category):
        info_list = dict()
        title = ia.get_movie(name)
        result = title.data["cover url"].find("._V1_")
        cover = title.data["cover url"].replace(
            title.data["cover url"][result + 3 : result + 23], ""
        )
        torrent_search = "tt" + name
        titles = Titles.get_magnet(torrent_search, category)
        tries = 0

        while True:
            if not titles and tries <= 5:
                time.sleep(2)
                titles = Titles.get_magnet(torrent_search, category)
                tries += 1
            elif not titles and tries >= 5:
                return "", 404
            else:
                tries = 0
                break

        script_dir = os.path.dirname(__file__)
        abs_path = os.path.join(script_dir, "log")
        if not os.path.exists(abs_path):
            os.makedirs(abs_path)
        with open(abs_path + "/magnets.txt", "a") as f:
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

        return info_list, cover

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
