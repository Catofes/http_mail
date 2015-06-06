__author__ = 'herbertqiao'

import falcon
import utils
from error import RError
from database import RDataBase


class RInvites():
    def __init__(self):
        self.db = RDataBase()

    @utils.require_codes
    def on_get(self, req, resp):
        resp.status = falcon.HTTP_200

