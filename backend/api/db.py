import time
import os
import mysql.connector
from mysql.connector import Error
from hashlib import sha256

# MySQL configurations
HOST = os.getenv("MYSQL_HOST")
PORT = os.getenv("MYSQL_PORT")
USER = os.getenv("MYSQL_USER")
PASSWORD = os.getenv("MYSQL_PASSWORD")
DATABASE = os.getenv("MYSQL_DATABASE")


def create_connection(host_name, user_name, user_password, db_name=None):
    try:
        connection = mysql.connector.connect(
            host=host_name, user=user_name, passwd=user_password, database=db_name
        )
        return connection
    except Error as e:
        print(f"The error '{e}' occurred")
        return None


def execute_query(query):
    connection = create_connection(HOST, USER, PASSWORD, DATABASE)
    cursor = connection.cursor()
    cursor.execute(query)
    connection.commit()
    cursor.close()
    connection.close()
    print("Query executed successfully")


# Connect to MySQL Server
while True:
    try:
        connection = create_connection(HOST, USER, PASSWORD)
        connection.close()
        break
    except Exception as e:
        print(e)
        time.sleep(1)

create_users_table = """
CREATE TABLE IF NOT EXISTS Users (
    UserID INT AUTO_INCREMENT PRIMARY KEY,
    Username VARCHAR(255) UNIQUE NOT NULL,
    Password VARCHAR(255)
);
"""

create_content_table = """
CREATE TABLE IF NOT EXISTS Content (
    title VARCHAR(255),
    description VARCHAR(255),
    url VARCHAR(255)
);
"""

# Creating tables
execute_query(create_users_table)
execute_query(create_content_table)


def initiate_db():
    connection = create_connection(HOST, USER, PASSWORD, DATABASE)
    cursor = connection.cursor()
    cursor.execute("SELECT Username FROM Users")
    myresult = cursor.fetchall()

    if not any("admin" in sl for sl in myresult):
        sql = "INSERT INTO Users (Username, Password) VALUES (%s, %s)"
        val = (
            "admin",
            "ec0d4ad57bd39659ea7263c15e44bb0e8b6628c3e087ae7f2ff0c1d216bf5da4",
        )
        cursor.execute(sql, val)
        connection.commit()

    cursor.execute("SELECT title FROM Content")
    myresult = cursor.fetchall()

    if not any("flag" in sl for sl in myresult):
        sql = "INSERT INTO Content (title, description, url) VALUES (%s, %s, %s)"
        val = [
            (
                "flag",
                "flag{you_sweaty_nerd}",
                "https://media1.tenor.com/m/XTII-SAYlPQAAAAd/enby-nonbinary.gif",
            ),
            (
                "Me when skatt",
                "Skatteetaten when 10 øre short on tax",
                "https://media1.tenor.com/m/0iCe8g03tcQAAAAC/fbi-fbi-open-up.gif",
            ),
            (
                "Shotscape 101",
                "Shotscape explaining why he has 0 points this ctf (he hosted it)",
                "https://media1.tenor.com/m/6RvyvMjx3XMAAAAC/he-is-speaking-guy-explaining-with-a-whiteboard.gif",
            ),
            (
                "Bendiks graduation",
                "(he's old)",
                "https://media1.tenor.com/m/1-kOdX4s8GsAAAAd/funny-birthday.gif",
            ),
        ]
        cursor.executemany(sql, val)
        connection.commit()
    cursor.close()
    connection.close()


initiate_db()


#################################
#################################
######### Not init ##############
#################################
#################################


def get_user_pass_auth(username, password_raw):
    connection = create_connection(HOST, USER, PASSWORD, DATABASE)
    cursor = connection.cursor()
    password = sha256(password_raw.encode("utf-8")).hexdigest()
    cursor.execute(
        "SELECT * FROM Users WHERE Username = %s AND Password = %s",
        (username, password),
    )
    record = cursor.fetchone()
    cursor.close()
    connection.close()
    return record


def get_description():
    connection = create_connection(HOST, USER, PASSWORD, DATABASE)
    cursor = connection.cursor()
    cursor.execute(
        "SELECT * FROM Content",
    )
    record = cursor.fetchall()
    cursor.close()
    connection.close()
    return record
