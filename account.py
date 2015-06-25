__author__ = 'herbertqiao'

from admin import RAdminUser
from database import RDataBase
import utils
from error import RError
import re


class RLogin:
    def __init__(self):
        pass

    @utils.require_login
    def on_get(self, req, resp, user):
        return

    def on_post(self, req, resp):
        if 'request' not in req.context.keys():
            raise RError(16)
        request = req.context['request']
        if 'username' not in request.keys():
            raise RError(20)
        if 'password' not in request.keys():
            raise RError(20)
        user = RAdminUser()
        token = utils.generate_code(16)
        user.login_by_password(request['username'], request['password'], token)
        req.context['result'] = {'token': token}

    @utils.require_login
    def on_delete(self, req, resp, user):
        user.logout()

    @utils.require_codes(1)
    def on_put(self, req, resp, code):
        user = RAdminUser()
        if 'request' not in req.context.keys():
            raise RError(16)
        request = req.context['request']
        if 'password' not in request.keys():
            raise RError(20)
        user.reset_password(code, request['password'])


class RRegister:
    def __init__(self):
        pass

    @utils.require_codes(0)
    def on_post(self, req, resp, code):
        if 'request' not in req.context.keys():
            raise RError(16)
        request = req.context['request']
        if 'username' not in request.keys():
            raise RError(20)
        if 'password' not in request.keys():
            raise RError(20)
        if not re.match('^[a-z0-9A-Z_]{3,60}$', request['username']):
            raise RError(13)
        user = RAdminUser()
        user.register(code, request['username'], request['password'])