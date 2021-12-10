from flask import Flask, render_template

app = Flask(__name__)


@app.route('/')
def json():
    return render_template('index.html')


def increment(temp):
    temp = 1
    return temp


@app.route('/background_process_test')
def background_process_test():
    return ("nothing")


if __name__ == '__main__':
    app.run()
