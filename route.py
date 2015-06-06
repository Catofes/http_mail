__author__ = 'herbertqiao'

import falcon
import middleware
from wsgiref import simple_server
from invites import RInvites
import domain
import user
app = falcon.API(middleware=[
    middleware.RequireJSON(),
    middleware.JSONTranslator()
]
                 )

invites = RInvites()
domains = domain.RDomain()
domainsModify = domain.RDomainModify()
users = user.RUser()
usersModify = user.RUserModify()

app.add_route('/invites', invites)
app.add_route('/domains', domains)
app.add_route('/domains/{domain_id}', domainsModify)
app.add_route('/user/{domain_id}', users)
app.add_route('/user/{domain_id}/{user_id}', usersModify)

if __name__ == '__main__':
    httpd = simple_server.make_server('127.0.0.1', 8000, app)
    httpd.serve_forever()
