__author__ = 'herbertqiao'

import utils
from database import RDataBasePool
from error import RError
import re


class RTransport:
    def __init__(self):
        self.db = RDataBasePool()

    @utils.require_login
    @utils.require_domain_owner
    def on_get(self, req, resp, user, domain_id):
        result = self.db.query("SELECT id,source,destination,region FROM transport_domains WHERE domain_id=%s",
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
        if not (re.match('^[a-z0-9A-Z_\+\.]{1,60}@$', request['source']) or request['source'] == ""):
            raise RError(13)
        domain = self.db.query("SELECT * FROM virtual_domains WHERE id = %s", (domain_id,))
        self.db.execute("INSERT INTO transport_domains(domain_id, source, destination, region) VALUES (%s,%s,%s,%s)",
                        (domain_id, (request['source'] + domain[0]['name']), request['destination'],
                         request['region']))


class RTransportModify:
    def __init__(self):
        self.db = RDataBasePool()

    @utils.require_login
    @utils.require_domain_owner
    def on_get(self, req, resp, user, domain_id, transport_id):
        result = self.db.query(
            "SELECT id,source,destination,region FROM transport_domains WHERE id = %s AND domain_id = %s",
            (transport_id, domain_id))
        if not result:
            raise RError(33)
        req.context['result'] = {"result": result[0]}

    @utils.require_login
    @utils.require_domain_owner
    @utils.require_level(5)
    def on_delete(self, req, resp, user, domain_id, transport_id):
        result = self.db.execute("DELETE FROM transport_domains WHERE id = %s AND domain_id = %s",
                                 (transport_id, domain_id))
        if not result:
            raise RError(18)


class RTransportDefault:
    def __init__(self):
        self.db = RDataBasePool()

    @utils.require_login
    def on_get(self, req, resp, user):
        result = {
            1: {
                "Illustrate": "Relay all mail to one server and store in it. This is helpful when you want to add multi "
                              "servers in your mx record.",
                "parameters": 'You need a json context with you server id in it. Like {"id":1}'
            }
        }
        req.context['result'] = {"result": result}


class RTransportDefaultModify:
    def __init__(self):
        self.db = RDataBasePool()

    @utils.require_login
    @utils.require_domain_owner
    def on_post(self, req, resp, user, domain_id, operate_id):
        if operate_id == "1":
            if 'request' not in req.context.keys():
                raise RError(16)
            request = req.context['request']
            if 'id' not in request.keys():
                raise RError(20)
            server = self.db.query("SELECT * FROM mynetworks WHERE id = %s", (str(request['id']),))
            domain = self.db.query("SELECT * FROM virtual_domains WHERE id = %s", (domain_id,))
            if not server:
                raise RError(30)
            cursor = self.db.begin()
            cursor.execute("DELETE FROM transport_domains WHERE domain_id = %s", (domain_id,))
            # cursor.execute(
            #    "INSERT INTO transport_domains(domain_id, source, destination, region) VALUES (%s, %s, %s, '0default')",
            #    (domain_id, domain[0]['name'], "smtp:[" + server[0]['domain_name'] + "]"))
            cursor.execute(
                "INSERT INTO transport_domains(domain_id, source, destination, region) VALUES( %s, %s, %s, %s)",
                (domain_id, domain[0]['name'], "lmtp:unix:private/dovecot-lmtp", "0default"))
            cursor.commit()
        else:
            raise RError(31)
