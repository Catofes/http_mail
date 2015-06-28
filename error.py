__author__ = 'herbertqiao'
import falcon
from falcon import HTTPError

Error_text = {
    0: ['Unknown Error',
        falcon.HTTP_500],
    1: ['SQL Error',
        falcon.HTTP_500],
    2: ['Json Required',
        falcon.HTTP_400],
    3: ['Empty Request Body',
        falcon.HTTP_400],
    4: ['Malformed JSON',
        falcon.HTTP_400],
    5: ['Required Invite Code',
        falcon.HTTP_400],
    6: ['Error Code',
        falcon.HTTP_404],
    7: ['Domain Required',
        falcon.HTTP_400],
    8: ['Domain Duplicated',
        falcon.HTTP_403],
    9: ['Domain Illegal',
        falcon.HTTP_403],
    10: ['Domain Id Required',
         falcon.HTTP_400],
    11: ['Error Code or Domain Id',
         falcon.HTTP_400],
    13: ['User Illegal',
         falcon.HTTP_403],
    14: ['Password Required',
         falcon.HTTP_400],
    15: ['User ID Required',
         falcon.HTTP_400],
    16: ['Empty Request',
         falcon.HTTP_403],
    17: ['User Duplicated',
         falcon.HTTP_403],
    18: ["Nothing Happened",
         falcon.HTTP_403],
    19: ["Username Duplicated",
         falcon.HTTP_400],
    20: ["Some Parameter is Missing",
         falcon.HTTP_400],
    21: ["Require Token",
         falcon.HTTP_400],
    22: ["Login Required",
         falcon.HTTP_403],
    23: ["Domain Id Required",
         falcon.HTTP_400],
    24: ["Domain Not Exist",
         falcon.HTTP_404],
    25: ["Error Username or Password",
         falcon.HTTP_400],
    26: ["User Not Exist",
         falcon.HTTP_404],
    27: ["Permission Deny",
         falcon.HTTP_403],
    28: ["Email Address Illegal",
         falcon.HTTP_400],
    29: ["Alias Not Exist",
         falcon.HTTP_400],
}


class RError(HTTPError):
    def __init__(self, code=0):
        global Error_text
        self.code = code
        if self.code not in Error_text.keys():
            self.code = 0
        self.text = Error_text[self.code][0]
        self.http_code = Error_text[self.code][1]
        HTTPError.__init__(self, self.http_code, self.text, code=self.code)