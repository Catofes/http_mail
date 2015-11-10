# -*- coding: utf-8 -*-
__author__ = 'herbertqiao'
from session import RMemorySessionStore
import time
from config import RConfig
from database import RDataBase
import hashlib
from error import RError
import re


class RAdminUserInfo():
    def __init__(self):
        self.id = 0
        self.username = ""
        self.token = ""
        self.ifLogin = False
        self.level = 0
        self.invite_code_id = 0
        self.expired_time = time.time() + 3600 * 24


class RAdminUser():
    def __init__(self, token=None):
        self.session = RMemorySessionStore()
        self.db = RDataBase()
        self.info = RAdminUserInfo()
        self.config = RConfig()
        if token:
            self.info.token = token
        if self.info.token != "":
            self.login_by_token()

    def login_by_token(self):
        if not self.info.token:
            return False
        if self.session.contains(self.info.token):
            self.info = self.session.get(self.info.token)
            if self.info.expired_time < time.time():
                self.info.ifLogin = False
                self.session.remove(self.info.token)
            if self.info.ifLogin:
                return True
        return False

    def login_by_password(self, username="", password="", token=""):
        if username == "" or password == "" or token == "":
            raise RError(20)
        password = hashlib.sha512(password + self.config.password_salt).hexdigest()[0: 64]
        result = self.db.query("SELECT * FROM admin_user WHERE username=%s AND password=%s", (username, password))
        if not result:
            raise RError(25)
        self.info.id = result[0]['id']
        self.info.ifLogin = True
        self.info.token = token
        return self.sync_from_mysql()

    def logout(self):
        if not self.info.ifLogin:
            return True
        if self.info.token == "":
            return True
        self.session.remove(self.info.token)
        self.info.ifLogin = False
        return True

    def sync_from_mysql(self):
        if not self.info.ifLogin:
            return False
        if self.info.id <= 0:
            return False
        result = self.db.query("SELECT * FROM admin_user WHERE  id = %s", (self.info.id,))
        if not result:
            return False
        self.info.username = result[0]['username']
        self.info.invite_code_id = result[0]['invite_code_id']
        self.info.level = result[0]['level']
        self.session.push(self.info.token, self.info)
        return True

    def register(self, code="", username="", password=""):
        if code == "" or password == "" or username == "":
            raise RError(20)
        password = hashlib.sha512(password + self.config.password_salt).hexdigest()[0: 64]
        if self.db.query("SELECT * FROM admin_user WHERE username = %s", (username,)):
            raise RError(19)
        result = self.db.execute(
            "INSERT INTO admin_user(username, password, level, invite_code_id) "
            "VALUES(%s,%s,1, (SELECT id FROM invite_codes WHERE code = %s))", (username, password, code))
        if not result:
            raise RError(0)
        result = self.db.execute("UPDATE invite_codes SET used = 1 WHERE code = %s", (code,))
        return True

    def reset_password(self, code="", password=""):
        if code == "" or password == "":
            raise RError(20)
        password = hashlib.sha512(password + self.config.password_salt).hexdigest()[0: 64]
        result = self.db.execute(
            "UPDATE admin_user SET password = %s WHERE invite_code_id in (SELECT id FROM invite_codes WHERE code = %s)",
            (password, code))
        return True
