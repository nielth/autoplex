from datetime import datetime
import json

def download_log(magnet_link, category):
    with open('./log/downloaded.json', 'r+') as file:
        logged_download = {
            "time": datetime.today().strftime('%H:%M:%S'),
            "category": category,
            "title": magnet_link['title'],
            "magnet": magnet_link['magnet']
        }
        try:
            file_data = json.load(file)
            file_data[datetime.today().strftime('%Y-%m-%d')
                    ].append(logged_download)
            file.seek(0)
            json.dump(file_data, file, indent=4)
        except json.decoder.JSONDecodeError:
            logged_download_new = {
                datetime.today().strftime('%Y-%m-%d'): [logged_download],
            }
            json.dump(logged_download_new, file, indent=4)
