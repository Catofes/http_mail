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
    

###Servers

Show all the mail servers in the system. 

    #List all servers
    GET:    /server?token=b5f1147824e37b8b
    RESP:   {"result": [{"default_mark": "0default", "server_mark": "CNAL", "domain_name": "a.b.com", "region_mark": "3CN"}]
    

###DKIM Settings

You can bind your dkim key to a domain. You can add one or replace one. The first number in this api is domain id. 
When you query your record, not key but sha512(key) will returned.

    #List a dkim record
    GET:    /dkim/1?token=b5f1147824e37b8b
    RESP:   {"key_sha512": "17979e1de7dc2574cc2113a452871e155c78997bd90ae4d03e86ee3d1b210938bc8f0c4f046e0fd715140d026d59c093e6e28a89f2dbed11b1fc3a426e1e832f", "domain": "a.com", "selector": "tau"}

    #Put a dkim record. If you already have one, this api will replace it.
    PUT:    /dkim/1?token=b5f1147824e37b8b
    BODY:   {"selector":"ppp", "private_key":"xxx"}
    
   
###BCC Settings

Show the BCC settings of the mail server. You need level upon 5. Please read postfix manual before add records. 
The region means which servers this record applies to. You can use server mark, region mark or default mark in this field. 
You need put your email username(without @domain.tld) in source field and put a intact email address in destination field.
If you put nothing in your username. All mails sent to your domain will be bcc to your destination.
The first number in the url is domain id and the second one is bcc record id.

    #List all bcc settings
    GET:    /bcc/1?token=b5f1147824e37b8b
    RESP:   {"result": [{"source": "r@aaa.com", "region": "SFDO", "destination": "r+relaycn@aaa.com", "id": 1}]}
    
    #Add a bcc settings
    POST:   /bcc/1?token=b5f1147824e37b8b
    BODY:   {"source":"abc","destination":"bb@gmail.com","region":"SFDO"}
    
    #List a bcc settings
    GET:    /bcc/1/4?token=f0872cbdc28b173
    RESP:   {"result": {"source": "abc@aaa.com", "region": "SFDO", "destination": "bb@gmail.com", "id": 4}}
    
    #Delete a bcc settings
    DELETE: /bcc/1/4?token=f0872cbdc28b173
    
    
###Alias Settings

Shows and modifies alias settings of a domain. Need level 5 above. 
The first number in the url is domain id and the second one is alias record id.

    #List all alias settings
    GET:    /alias/1?token=b5f1147824e37b8b
    RESP:   {"result": [{"source": "sdf@aaa.com", "destination": "haha@gmail.com", "id": 8}]}
    
    #Add a alias settings
    POST:   /alias/1?token=f0872cbdc28b173
    BODY:   {"source":"sdf","destination":"haha@gmail.com"}
    
    #List a alias settings
    GET:    /alias/1/8?token=f0872cbdc28b173
    RESP:   {"result": {"source": "sdf@aaa.com", "destination": "haha@gmail.com", "id": 8}}
    
    #Delete a alias settings
    DELETE: /alias/1/8?token=f0872cbdc28b173

###Transport Settings

Shows and modifies transport settings of a domain. Please read postfix transport manual.
Need level 5 above. You need put "username@" in source field or put nothing in it.

    #List all transport settings
    GET:    /transport/2?token=b5f1147824e37b8b
    RESP:   {"result": [{"source": "aaa.com", "region": "SFDO", "destination": "lmtp:unix:private/dovecot-lmtp", "id": 4}, {"source": "aaa.com", "region": "0default", "destination": "smtp:[xxx.domain.tld]", "id": 5}]}
    
    #Add a transport settings
    POST:   /transport/2?token=f0872cbdc28b173
    BODY:   {"source":"k@","destination":"smtp:[smtp.google.com]","region":"0default"}
    
    #List a transport settings
    GET:    /transport/2/14?token=f0872cbdc28b173
    RESP:   {"result": {"source": "k@aaa.com", "region": "0default", "destination": "smtp:[smtp.google.com]", "id": 14}}
    
    #Delete a transport settings
    DELETE: /transport/2/14?token=f0872cbdc28b173
    
####Default Transport Settings

Add some default transport settings. This will delete all your transport settings of a domain. Be careful.

    #Get what you can do 
    GET:    /transport_default
    
    #Do Something. First number is domain id and second number is operate id.
    POST:   /transport_default/1/1?token=f0872cbdc28b173