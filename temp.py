import requests

from markupsafe import escape
from flask import Flask, jsonify, request

app = Flask(__name__)


@app.route("/api/tvmaze/search/<search>", methods=["GET"])
def display_data(search):
    search = escape(search)
    s_path: str= f"https://api.tvmaze.com/singlesearch/shows?q={search}"
    try:
        resp = requests.get(s_path).json()
    except Exception as e:
        return "", 500

    return resp


if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=5051)
