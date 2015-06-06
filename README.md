##Required

python 2.7
falcon
DBUtils
MySQLdb

NOT WORKED IN PYTHON 3

##RUN

python route.py

##API

Every request need a url params "code" meaning invite code. 

###invites

url: /invites?code=123

#### method: GET

Check if invite code is right. If the code works it will return HTTP CODE at 200

###domains

url: /domains?code=123

####method: GET

List all domains bind to the invite code.

response example:

    {"result": [{"id": 9, "name": "aaa.com"}, {"id": 10, "name": "sfs.com"}]}


####method: POST

Create a domain and bind it to the invite code.

request example:

	{"domain": "aaa.com"}

response example:

	{"result": [{"id": 11, "name": "aaa.com"}]}

###user

url: /user/{domain_id}?code=123

####method: GET

List all user bind to the domain.

response example:

	{"result": [{"email": "adas@aaa.com", "id": 15, "domain_id": 9}]}

####method: POST

Add a user to the domain.

request example:

	{
		    "user": "adsas",
			"password": "tatata"
	}

HTTP CODE 200 will returned if everything worked.

###user modify

url: /user/{domain_id}/{user_id}

####method: GET

show the user.

response example:

	{"result": [{"email": "adas@aaa.com", "id": 15, "domain_id": 9}]}

####method: PUT

change the user's password.

request example:
	
	{"password": "sss"}

HTTP CODE 200 will returned if everything worked.

####method: DELETE

delete the user.

HTTP CODE 200 will returned if everything worked.


