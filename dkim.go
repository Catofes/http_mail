package main

import (
	"crypto/sha512"
	"encoding/hex"
)

func (a *api) dkim(r *request) (any, error) {
	var obj map[string]any
	var err error
	if r.Method == "PUT" {
		obj, err = r.object()
		if err != nil {
			return nil, err
		}
		if err := requireFields(obj, "selector", "private_key"); err != nil {
			return nil, err
		}
	}
	domain, err := a.domainRow(r)
	if err != nil {
		return nil, err
	}
	signings, err := a.db.query(r.Context(), "SELECT * FROM opendkim_signings WHERE author = $1", domain["name"])
	if err != nil {
		return nil, err
	}
	if r.Method == "GET" {
		if len(signings) == 0 {
			return nil, nil
		}
		keys, err := a.db.query(r.Context(), "SELECT * FROM opendkim_keys WHERE id = $1", signings[0]["dkim_id"])
		if err != nil {
			return nil, err
		}
		if len(keys) == 0 {
			return nil, nil
		}
		hash := sha512.Sum512([]byte(keys[0]["private_key"].(string)))
		return map[string]any{"domain": keys[0]["domain_name"], "selector": keys[0]["selector"], "key_sha512": hex.EncodeToString(hash[:])}, nil
	}
	if len(signings) == 0 {
		if err := a.db.execute(r.Context(), "INSERT INTO opendkim_keys(domain_name, selector, private_key) VALUES ($1,$2,$3)",
			domain["name"], obj["selector"], obj["private_key"]); err != nil {
			return nil, err
		}
		return nil, a.db.execute(r.Context(), "INSERT INTO opendkim_signings(author, dkim_id) VALUES ($1, (SELECT id FROM opendkim_keys WHERE domain_name = $2))", domain["name"], domain["name"])
	}
	return nil, a.db.execute(r.Context(), "UPDATE opendkim_keys SET selector=$1, private_key=$2 WHERE domain_name = $3", obj["selector"], obj["private_key"], domain["name"])
}
