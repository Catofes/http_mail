__author__ = 'herbertqiao'

import falcon
import middleware
from wsgiref import simple_server
from invites import RInvites
from server import RServer
from dkim import RDKIM
import domain
import user
import account
import alias
import bcc
import transport

app = falcon.API(middleware=[
    middleware.RequireJSON(),
    middleware.JSONTranslator()
]
                 )

r_invites = RInvites()
r_domains = domain.RDomain()
r_domainsModify = domain.RDomainModify()
r_users = user.RUser()
r_usersModify = user.RUserModify()
r_login = account.RLogin()
r_register = account.RRegister()
r_server = RServer()
r_dkim = RDKIM()
r_alias = alias.RAlias()
r_aliasModify = alias.RAliasModify()
r_bcc = bcc.RBcc()
r_bccModify = bcc.RBccModify()
r_transport = transport.RTransport()
r_transportModify = transport.RTransportModify()
r_transportDefault = transport.RTransportDefault()
r_transportDefaultModify = transport.RTransportDefaultModify()

app.add_route('/invite', r_invites)
app.add_route('/login', r_login)
app.add_route('/register', r_register)
app.add_route('/server', r_server)
app.add_route('/domain', r_domains)
app.add_route('/domain/{domain_id}', r_domainsModify)
app.add_route('/user/{domain_id}', r_users)
app.add_route('/user/{domain_id}/{user_id}', r_usersModify)
app.add_route('/dkim/{domain_id}', r_dkim)
app.add_route('/alias/{domain_id}', r_alias)
app.add_route('/alias/{domain_id}/{alias_id}', r_aliasModify)
app.add_route('/bcc/{domain_id}', r_bcc)
app.add_route('/bcc/{domain_id}/{bcc_id}', r_bccModify)
app.add_route('/transport/{domain_id}', r_transport)
app.add_route('/transport/{domain_id}/{transport_id}', r_transportModify)
app.add_route('/transport_default', r_transportDefault)
app.add_route('/transport_default/{domain_id}/{operate_id}', r_transportDefaultModify)

if __name__ == '__main__':
    httpd = simple_server.make_server('0.0.0.0', 8000, app)
    httpd.serve_forever()
