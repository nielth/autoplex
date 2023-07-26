import sqlite3

connect = sqlite3.connect('data.db')
connect.execute(
    'CREATE TABLE IF NOT EXISTS PARTICIPANTS (username TEXT, \
    email TEXT, city TEXT, country TEXT, phone TEXT)')