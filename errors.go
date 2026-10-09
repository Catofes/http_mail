package main

type legacyError struct {
	Code   int    `json:"code"`
	Title  string `json:"title"`
	status int
}

func (e *legacyError) Error() string { return e.Title }

var errorTexts = map[int]struct {
	text   string
	status int
}{
	0: {"Unknown Error.", 500}, 1: {"SQL Error.", 500},
	2: {"Json Required.", 400}, 3: {"Empty Request Body.", 400}, 4: {"Malformed JSON.", 400},
	5: {"Required Invite Code.", 400}, 6: {"Error Invite Code.", 404},
	7: {"Domain Required.", 400}, 8: {"Domain Duplicated.", 403}, 9: {"Domain Illegal.", 403},
	10: {"Domain Id Required.", 400}, 11: {"Error Code or Domain Id.", 400},
	13: {"Username Illegal.", 403}, 14: {"Password Required.", 400}, 15: {"User ID Required.", 400},
	16: {"Empty Request. Parameter Need.", 403}, 17: {"User Duplicated.", 403},
	18: {"Nothing Happened. The Database was not Changed.", 403}, 19: {"Username Duplicated.", 400},
	20: {"Some Parameter is Missing.", 400}, 21: {"Require Token.", 400}, 22: {"Login Required.", 403},
	23: {"Domain Id Required.", 400}, 24: {"Domain Not Exist.", 404}, 25: {"Error Username or Password.", 400},
	26: {"User Not Exist.", 404}, 27: {"Permission Deny.", 403}, 28: {"Email Address Illegal.", 400},
	29: {"Alias Not Exist.", 404}, 30: {"Error Server Id.", 400}, 31: {"Request Error. Operate Not Found.", 400},
	32: {"BCC Not Exist.", 404}, 33: {"Transport Not Exist.", 404},
}

func apiError(code int) *legacyError {
	t, ok := errorTexts[code]
	if !ok {
		code, t = 0, errorTexts[0]
	}
	return &legacyError{Code: code, Title: t.text, status: t.status}
}
