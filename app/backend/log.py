from datetime import datetime
import json
import os


def download_log(magnet, title, category):
    script_dir = os.path.dirname(__file__)
    abs_path = os.path.join(script_dir, "log")
    if not os.path.exists(abs_path):
        os.makedirs(abs_path)
    with open(abs_path + "/downloaded.json", "a") as file:
        logged_download = {
            "time": datetime.today().strftime("%H:%M:%S"),
            "category": category,
            "title": title,
            "magnet": magnet,
        }
        try:

            file_data = json.load(file)
            file_data[datetime.today().strftime("%Y-%m-%d")].append(logged_download)
            file.seek(0)
            json.dump(file_data, file, indent=4)
        except:
            logged_download_new = {
                datetime.today().strftime("%Y-%m-%d"): logged_download
            }
            json.dump(logged_download_new, file, indent=4)
