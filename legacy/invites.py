__author__ = 'herbertqiao'

import falcon
import utils


class RInvites():
    def __init__(self):
        pass

    @utils.require_codes(0)
    def on_get(self, req, resp, code):
        resp.status = falcon.HTTP_200
