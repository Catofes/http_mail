package main

import (
	"errors"
	"regexp"
	"strings"
)

var (
	adminName       = regexp.MustCompile(`^[a-z0-9A-Z_]{3,60}$`)
	mailboxName     = regexp.MustCompile(`^[a-z0-9A-Z][a-z0-9A-Z_+.]{0,60}$`)
	ruleSource      = regexp.MustCompile(`^[a-z0-9A-Z_+.]{0,60}$`)
	transportSource = regexp.MustCompile(`^[a-z0-9A-Z_+.]{1,60}@$`)
	domainName      = regexp.MustCompile(`^([a-zA-Z0-9][-a-zA-Z0-9]{0,62}\.){1,5}[a-zA-Z0-9][-a-zA-Z0-9]{0,62}$`)
	emailAddress    = regexp.MustCompile(`^[a-z0-9A-Z_+.]{1,60}@([a-zA-Z0-9][-a-zA-Z0-9]{0,62}\.){1,5}[a-zA-Z0-9][-a-zA-Z0-9]{0,62}$`)
)

func matches(pattern *regexp.Regexp, value string) bool {
	// Python's $ also matches immediately before a final newline.
	return pattern.MatchString(strings.TrimSuffix(value, "\n"))
}

func (a *api) domainRow(r *request) (row, error) {
	rows, err := a.db.query(r.Context(), "SELECT * FROM virtual_domains WHERE id = $1", r.parts[1])
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("missing domain")
	}
	return rows[0], nil
}

func (a *api) domain(r *request) (any, error) {
	if len(r.parts) == 2 {
		if r.Method == "GET" {
			return a.one(r, 24, "SELECT id, name FROM virtual_domains WHERE id = $1", r.parts[1])
		}
		if err := a.db.execute(r.Context(), "DELETE FROM virtual_users WHERE domain_id = $1", r.parts[1]); err != nil {
			return nil, err
		}
		return nil, a.db.execute(r.Context(), "DELETE FROM virtual_domains WHERE id = $1", r.parts[1])
	}
	if r.Method == "GET" {
		if r.user.level == 100 {
			return a.list(r, "SELECT id, name FROM virtual_domains")
		}
		return a.list(r, "SELECT id, name FROM virtual_domains WHERE admin_user_id = $1", r.user.id)
	}
	obj, err := r.object()
	if err != nil {
		return nil, err
	}
	name, err := stringField(obj, "domain", 7)
	if err != nil {
		return nil, err
	}
	rows, err := a.db.query(r.Context(), "SELECT id FROM virtual_domains WHERE name = $1", name)
	if err != nil {
		return nil, err
	}
	if len(rows) != 0 {
		return nil, apiError(8)
	}
	if !matches(domainName, name) {
		return nil, apiError(9)
	}
	if err := a.db.execute(r.Context(), "INSERT INTO virtual_domains (name, admin_user_id) VALUES ($1, $2)", name, r.user.id); err != nil {
		return nil, err
	}
	rows, err = a.db.query(r.Context(), "SELECT id, name FROM virtual_domains WHERE name = $1", name)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("inserted domain disappeared")
	}
	return map[string]any{"result": rows[0]}, nil
}
