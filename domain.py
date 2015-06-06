__author__ = 'herbertqiao'

import falcon
import utils
import re
from error import RError
from database import RDataBase


class RDomain:
    def __init__(self):
        self.db = RDataBase()

    @utils.require_codes
    def on_get(self, req, resp):
        code = req.params['code']
        result = self.db.query("SELECT id, name FROM virtual_domains WHERE code = %s", (code,))
        req.context['result'] = {'result': result}
        resp.status = falcon.HTTP_200


    @utils.require_codes
    def on_post(self, req, resp):
        code = req.params['code']
        if 'request' not in req.context.keys():
            raise RError(16)
        request = req.context['request']
        if 'domain' not in request.keys():
            raise RError(7)
        if self.db.query("SELECT id FROM virtual_domains WHERE name = %s", (request['domain'],)):
            raise RError(8)
        if not re.match('^[a-zA-Z0-9][-a-zA-Z0-9]{0,62}\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62}$', request['domain']):
            raise RError(9)
        self.db.execute("INSERT INTO virtual_domains (name, code) VALUES (%s, %s)", (request['domain'], code))
        req.context['result'] = {'result': self.db.query('SELECT id, name FROM virtual_domains WHERE name = %s',
                                                         (request['domain'],))}
        resp.status = falcon.HTTP_200







