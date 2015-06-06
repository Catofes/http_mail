__author__ = 'herbertqiao'

from error import RError
from database import RDataBase


def require_codes(func):
    def check_login(self, req, resp):
        if 'code' not in req.params.keys():
            raise RError(5)
        db = RDataBase()
        if not db.query("SELECT * FROM invite_codes WHERE code = %s AND used = 1", (req.params['code'],)):
            raise RError(6)
        func(self, req, resp)

    return check_login


def require_codes_and_domain(func):
    def check_login(self, req, resp, **kwargs):
        if 'code' not in req.params.keys():
            raise RError(5)
        db = RDataBase()
        if 'domain_id' not in kwargs.keys():
            raise RError(10)
        domain_id = kwargs['domain_id']
        if not db.query("SELECT * FROM virtual_domains WHERE id = %s AND code = %s",
                        (domain_id, req.params['code'],)):
            raise RError(11)
        func(self, req, resp, **kwargs)

    return check_login

