from datetime import datetime
import json
import os


def store_magnet(magnet):
    script_dir = os.path.dirname(__file__)
    abs_path = os.path.join(script_dir, "log/magnets.txt")
    legit_magnet = False
    for line in reversed(open(abs_path).readlines()):
        temp = line.rstrip()
        if line.rstrip() in magnet:
            legit_magnet = True
            break
    if not legit_magnet:
        return "", 404


def log_user_download(current_user, category, title, magnet):
    script_dir = os.path.dirname(__file__)
    with open(os.path.join(script_dir, "log/download.json"), "r+") as file_read:
        try:
            read_json = json.load(file_read)
        except json.decoder.JSONDecodeError:
            read_json = {}
        lst = {
            "title": title,
            "user": current_user.email,
            "time": datetime.now().strftime("%Y-%m-%d %H:%M"),
            "category": category,
            "magnet": magnet,
        }
        existing_key = read_json.get(datetime.now().strftime("%Y-%m-%d %H:%M"))
        if not existing_key:
            read_json[datetime.now().strftime("%Y-%m-%d %H:%M")] = {}
        read_json[datetime.now().strftime("%Y-%m-%d %H:%M")][title] = lst
        file_read.seek(0)
        json.dump(read_json, file_read, indent=4)


def get_user_download():
    script_dir = os.path.dirname(__file__)
    with open(os.path.join(script_dir, "log/download.json"), "r") as file_read:
        try:
            read_json = json.load(file_read)
        except json.decoder.JSONDecodeError:
            read_json = {}
        return read_json
