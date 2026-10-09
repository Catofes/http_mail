package main

func (a *api) mailbox(r *request) (any, error) {
	domainID := r.parts[1]
	if r.Method == "GET" {
		if len(r.parts) == 2 {
			return a.list(r, "SELECT id, domain_id, email FROM virtual_users WHERE domain_id = $1", domainID)
		}
		return a.one(r, 26, "SELECT id, domain_id, email FROM virtual_users WHERE id = $1 AND domain_id = $2", r.parts[2], domainID)
	}
	if r.Method == "DELETE" {
		return nil, a.db.execute(r.Context(), "DELETE FROM virtual_users WHERE id = $1 AND domain_id = $2", r.parts[2], domainID)
	}
	obj, err := r.object()
	if err != nil {
		return nil, err
	}
	var email string
	if r.Method == "POST" {
		username, err := stringField(obj, "username", 20)
		if err != nil {
			return nil, err
		}
		if !matches(mailboxName, username) {
			return nil, apiError(13)
		}
		if _, ok := obj["password"]; !ok {
			return nil, apiError(14)
		}
		domain, err := a.domainRow(r)
		if err != nil {
			return nil, err
		}
		email = username + "@" + domain["name"].(string)
		rows, err := a.db.query(r.Context(), "SELECT * FROM virtual_users WHERE email = $1", email)
		if err != nil {
			return nil, err
		}
		if len(rows) != 0 {
			return nil, apiError(17)
		}
	}
	password, err := stringField(obj, "password", 14)
	if err != nil {
		return nil, err
	}
	hash, err := mailboxPassword(password)
	if err != nil {
		return nil, err
	}
	if r.Method == "POST" {
		return nil, a.db.execute(r.Context(), "INSERT INTO virtual_users (domain_id, password, email) VALUES ($1, $2, $3)", domainID, hash, email)
	}
	return nil, a.db.execute(r.Context(), "UPDATE virtual_users SET password = $1 WHERE domain_id = $2 AND id = $3", hash, domainID, r.parts[2])
}
