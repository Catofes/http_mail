__author__ = 'herbertqiao'

import utils
from database import RDataBase
from error import RError
import re


class RBcc:
    def __init__(self):
        self.db = RDataBase()

    @utils.require_login
    @utils.require_domain_owner
    def on_get(self, req, resp, user, domain_id):
        result = self.db.query("SELECT id,source,destination,region FROM recipient_bcc WHERE domain_id = %s",
                               (domain_id,))
        req.context['result'] = {"result": result}

    @utils.require_login
    @utils.require_domain_owner
    @utils.require_level(5)
    def on_post(self, req, resp, user, domain_id):
        if 'request' not in req.context.keys():
            raise RError(16)
        request = req.context['request']
        if 'source' not in request.keys():
            raise RError(20)
        if 'destination' not in request.keys():
            raise RError(20)
        if 'region' not in request.keys():
            raise RError(20)
        if not re.match('^[a-z0-9A-Z_\+]{0,60}$', request['source']):
            raise RError(13)
        if not re.match('^[a-z0-9A-Z_\+]{1,60}@[a-zA-Z0-9][-a-zA-Z0-9]{0,62}\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62}$',
                        request['destination']):
            raise RError(28)
        domain = self.db.query("SELECT * FROM virtual_domains WHERE id = %s", (domain_id,))
        self.db.execute("INSERT INTO recipient_bcc(domain_id, source, destination, region) VALUES (%s,%s,%s,%s)",
                        (domain_id, (request['source'] + "@" + domain[0]['name']), request['destination'],
                         request['region']))


class RBccModify:
    def __init__(self):
        self.db = RDataBase()

    @utils.require_login
    @utils.require_domain_owner
    def on_get(self, req, resp, user, domain_id, bcc_id):
        result = self.db.query(
            "SELECT id,source,destination,region FROM recipient_bcc WHERE id = %s AND domain_id = %s",
            (bcc_id, domain_id))
        if not result:
            raise RError(32)
        req.context['result'] = {"result": result[0]}

    @utils.require_login
    @utils.require_domain_owner
    @utils.require_level(5)
    def on_delete(self, req, resp, user, domain_id, bcc_id):
        result = self.db.execute("DELETE FROM recipient_bcc WHERE id = %s AND domain_id = %s", (bcc_id, domain_id))
        if not result:
            raise RError(18)
