##Required

python 2.7
falcon
DBUtils
MySQLdb

NOT WORKED IN PYTHON 3

##RUN

python route.py

##API

###Common

The Request must encode in JSON format. And all the api will return JSON if there is something needed to be returned.

####Error Code:

If everything worked perfect, a HTTP 200 will returned. All other HTTP STATUS CODE means some errors. 
A error code and a illustrate will be returned. Error codes are shown below.

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
    }


###Invite

Check whether the invite code is available.

    GET /invite?code=123

###Register

Register a new user. You need send username and password in body with invite code in url parameter.

    POST:   /register?code=123
    BODY:   {"username":"catofes","password":"321"}
    
###Login

A token(16-char) will returned if you login successfully. The token will expired in 24h. You need your invite code to change your password.

    #Check if token available.
    GET:    /login?token=1e987d1eba781730
    
    #Login
    POST:   /login
    BODY:   {"username":"catofes","password":"321"}
    RESP:   {"token": "bd36d1ccb2884d6d","username":"catofes","level":1}

    #Logout
    DELETE: /login?token=1e987d1eba781730
    
    #ChangePassword
    PUT:    /login?code=123
    BODY:   {"password":"321"}

###Domains

You can list your domains, add domain or delete domain. When you list a domain or delete it, you need use your domain id instead of number 18.

    #List your domains
    GET:    /domain?token=865d54814424abbe
    RESP:   {"result": [{"id": 16, "name": "sssfsdd.com"}, {"id": 18, "name": "aadd.com"}]}
    
    #Add a domain
    POST:   /domain?token=865d54814424abbe
    BODY:   {"domain": "aadd.com"}
    RESP:   {"result": {"id": 18, "name": "aadd.com"}}
    
    #List a domain 
    GET:    /domain/18?token=865d54814424abbe
    RESP:   {"result": {"id": 18, "name": "aadd.com"}}
    
    #Delete a domain
    DELETE: /domain/18?token=865d54814424abbe
    

###Users

Modify email users belong to a domain. You can list users, add user, delete user or change user's password.
The first number in this api means domain id and the second one means user id.

    #List all user in a domain
    GET:    /user/18?token=865d54814424abbe
    RESP:   {"result": [{"email": "adas@aadd.com", "id": 20, "domain_id": 18}, {"email": "adasss@aadd.com", "id": 21, "domain_id": 18}]}
    
    #Add a user into a domain
    POST:   /user/18?token=865d54814424abbe
    BODY:   {"username": "adas","password": "tatata"}
    
    #List a user
    GET:    /user/18/20?token=865d54814424abbe
    RESP:   {"result": {"email": "adas@aadd.com", "id": 20, "domain_id": 18}}
    
    #Change user's password
    PUT:    /user/18/20?token=865d54814424abbe
    BODY:   {"password": "123"}
    
    #Delete a user
    DELETE: /user/18/20?token=865d54814424abbe
    



