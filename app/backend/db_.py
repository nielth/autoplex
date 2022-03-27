from flask_sqlalchemy import SQLAlchemy
from app import db, UserMixin


class User(UserMixin, db.Model):
    id = db.Column(db.Integer, primary_key=True)
    email = db.Column(db.String(100), nullable=False, unique=True)
    uuid = db.Column(db.String(100), nullable=True, unique=True)
    username = db.Column(db.String(100), nullable=True, unique=True)
    password = db.Column(db.String(100), nullable=True, unique=True)


db.create_all()
