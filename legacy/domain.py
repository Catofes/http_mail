__author__ = 'herbertqiao'

import falcon
import utils
import re
from error import RError
from database import RDataBasePool


class RDomain:
    def __init__(self):
        self.db = RDataBasePool()

    @utils.require_login
    def on_get(self, req, resp, user):
        if user.info.level == 100:
            result = self.db.query("SELECT id, name FROM virtual_domains", ())
        else:
            result = self.db.query("SELECT id, name FROM virtual_domains WHERE admin_user_id = %s", (user.info.id,))
        if result == "":
            result = []
        req.context['result'] = {'result': result}
        resp.status = falcon.HTTP_200


    @utils.require_login
    def on_post(self, req, resp, user):
        if 'request' not in req.context.keys():
            raise RError(16)
        request = req.context['request']
        if 'domain' not in request.keys():
            raise RError(7)
        if self.db.query("SELECT id FROM virtual_domains WHERE name = %s", (request['domain'],)):
            raise RError(8)
        if not re.match('^([a-zA-Z0-9][-a-zA-Z0-9]{0,62}\.){1,5}[a-zA-Z0-9][-a-zA-Z0-9]{0,62}$', request['domain']):
            raise RError(9)
        self.db.execute("INSERT INTO virtual_domains (name, admin_user_id) VALUES (%s, %s)",
                        (request['domain'], user.info.id))
        req.context['result'] = {'result': self.db.query('SELECT id, name FROM virtual_domains WHERE name = %s',
                                                         (request['domain'],))[0]}
        resp.status = falcon.HTTP_200


class RDomainModify():
    def __init__(self):
        self.db = RDataBasePool()

    @utils.require_login
    @utils.require_domain_owner
    def on_get(self, req, resp, domain_id, user):
        result = self.db.query('SELECT id, name FROM virtual_domains WHERE id = %s;', (domain_id, ))
        if not result:
            raise RError(24)
        req.context['result'] = {'result': result[0]}
        resp.status = falcon.HTTP_200


    @utils.require_login
    @utils.require_domain_owner
    def on_delete(self, req, resp, domain_id, user):
        result = self.db.execute("DELETE FROM virtual_users WHERE domain_id = %s;", (domain_id,))
        result = self.db.execute("DELETE FROM virtual_domains WHERE id = %s;", (domain_id,))
        resp.status = falcon.HTTP_200






