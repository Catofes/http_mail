__author__ = 'herbertqiao'

import falcon
import utils
from database import RDataBase
import hashlib
from error import RError


class RDKIM:
    def __init__(self):
        self.db = RDataBase()

    @utils.require_login
    @utils.require_domain_owner
    def on_get(self, req, resp, user, domain_id):
        domain = self.db.query("SELECT * FROM virtual_domains WHERE id = %s", (domain_id,))
        opendkim_id = self.db.query("SELECT * FROM opendkim_signings WHERE author = %s", (domain[0]['name'],))
        if not opendkim_id:
            return
        opendkim_key = self.db.query("SELECT * FROM opendkim_keys WHERE  id =%s", (opendkim_id[0]['dkim_id'],))
        if not opendkim_key:
            return
        req.context['result'] = {"domain": opendkim_key[0]['domain_name'], "selector": opendkim_key[0]['selector'],

                                 "key_sha512": hashlib.sha512(opendkim_key[0]['private_key']).hexdigest()}

    @utils.require_login
    @utils.require_domain_owner
    def on_put(self, req, resp, user, domain_id):
        if 'request' not in req.context.keys():
            raise RError(16)
        request = req.context['request']
        if 'selector' not in request.keys():
            raise RError(20)
        if 'private_key' not in request.keys():
            raise RError(20)
        domain = self.db.query("SELECT * FROM virtual_domains WHERE id = %s", (domain_id,))
        opendkim_id = self.db.query("SELECT * FROM opendkim_signings WHERE author = %s", (domain[0]['name'],))
        if not opendkim_id:
            self.db.execute("INSERT INTO opendkim_keys(domain_name, selector, private_key) VALUES (%s,%s,%s)",
                            (domain[0]['name'], request['selector'], request['private_key']))
            self.db.execute(
                "INSERT INTO opendkim_signings(author, dkim_id) "
                "VALUES (%s, (SELECT id FROM opendkim_keys WHERE domain_name = %s))",
                (domain[0]['name'], domain[0]['name']))
        else:
            self.db.execute("UPDATE opendkim_keys SET selector=%s, private_key=%s WHERE domain_name = %s",
                            (request['selector'], request['private_key'], domain[0]['name']))
