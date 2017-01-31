__author__ = 'herbertqiao'

import falcon
import utils
from database import RDataBasePool


class RServer:
    def __init__(self):
        self.db = RDataBasePool()

    @utils.require_login
    def on_get(self, req, resp, user):
        result = self.db.query("SELECT domain_name, server_mark, region_mark, default_mark FROM mynetworks", ())
        req.context['result'] = {"result": result}


