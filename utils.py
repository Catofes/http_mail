__author__ = 'herbertqiao'

from error import RError
from database import RDataBase
from admin import RAdminUser
import os


def require_codes(used=0):
    def dec(func):
        def check_login(*args, **kwargs):
            if 'code' not in args[1].params.keys():
                raise RError(5)
            db = RDataBase()
            if not db.query("SELECT * FROM invite_codes WHERE code = %s AND used = %s", (args[1].params['code'], used)):
                raise RError(6)
            func(*args, code=args[1].params['code'], **kwargs)

        return check_login

    return dec


def require_login(func):
    def check_login(*args, **kwargs):
        if 'token' not in args[1].params.keys():
            raise RError(21)
        user = RAdminUser(token=args[1].params['token'])
        user.login_by_token()
        if not user.info.ifLogin:
            raise RError(22)
        func(*args, user=user, **kwargs)

    return check_login


def require_domain_owner(func):
    def check_owner(*args, **kwargs):
        if 'user' not in kwargs.keys():
            raise RError(22)
        user = kwargs['user']
        if 'domain_id' not in kwargs.keys():
            raise RError(23)
        if user.info.level == 100:
            return func(*args, **kwargs)
        db = RDataBase()
        if not db.query("SELECT * FROM virtual_domains WHERE id = %s AND admin_user_id = %s",
                        (kwargs['domain_id'], kwargs['user'].info.id)):
            raise RError(24)
        func(*args, **kwargs)

    return check_owner

def require_level(level=0):
    def dec(func):
        def check_login(*args, **kwargs):
            if 'user' not in kwargs.keys():
                raise RError(22)
            if kwargs['user'].info.level < level:
                raise RError(27)
            func(*args, **kwargs)

        return check_login

    return dec


def generate_code(length):
    if length < 1:
        return False
    length = length / 2
    return ''.join(map(lambda xx: (hex(ord(xx))[2:]), os.urandom(length)))


