package main

import (
	"fmt"
	"strconv"
	"time"
)

func requireFields(obj map[string]any, keys ...string) error {
	for _, key := range keys {
		if _, ok := obj[key]; !ok {
			return apiError(20)
		}
	}
	return nil
}

func (a *api) account(r *request, code string) (any, error) {
	obj, err := r.object()
	if err != nil {
		return nil, err
	}
	if r.Method == "PUT" {
		password, err := stringField(obj, "password", 20)
		if err != nil {
			return nil, err
		}
		if password == "" || code == "" {
			return nil, apiError(20)
		}
		return nil, a.db.execute(r.Context(),
			"UPDATE admin_users SET password = $1 WHERE invite_code_id in (SELECT id FROM invite_codes WHERE code = $2)",
			adminPassword(password, a.config.PasswordSalt), code)
	}
	if err := requireFields(obj, "username", "password"); err != nil {
		return nil, err
	}
	username, err := stringField(obj, "username", 20)
	if err != nil {
		return nil, err
	}
	if r.parts[0] == "register" && !matches(adminName, username) {
		return nil, apiError(13)
	}
	password, err := stringField(obj, "password", 20)
	if err != nil {
		return nil, err
	}
	if username == "" || password == "" {
		return nil, apiError(20)
	}
	hash := adminPassword(password, a.config.PasswordSalt)
	if r.parts[0] == "register" {
		rows, err := a.db.query(r.Context(), "SELECT * FROM admin_users WHERE username = $1", username)
		if err != nil {
			return nil, err
		}
		if len(rows) != 0 {
			return nil, apiError(19)
		}
		if err := a.db.execute(r.Context(), "INSERT INTO admin_users(username, password, level, invite_code_id) "+
			"VALUES($1,$2,1, (SELECT id FROM invite_codes WHERE code = $3))", username, hash, code); err != nil {
			return nil, err
		}
		return nil, a.db.execute(r.Context(), "UPDATE invite_codes SET used = 1 WHERE code = $1", code)
	}
	expires := time.Now().Add(24 * time.Hour)
	token, err := generateToken()
	if err != nil {
		return nil, err
	}
	rows, err := a.db.query(r.Context(), "SELECT * FROM admin_users WHERE username=$1 AND password=$2", username, hash)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apiError(25)
	}
	rows, err = a.db.query(r.Context(), "SELECT * FROM admin_users WHERE id = $1", rows[0]["id"])
	if err != nil {
		return nil, err
	}
	// Preserve the empty account response if the account vanished between reads.
	if len(rows) == 0 {
		return map[string]any{"token": token, "username": "", "level": 0}, nil
	}
	level, err := strconv.ParseInt(fmt.Sprint(rows[0]["level"]), 10, 64)
	if err != nil {
		return nil, err
	}
	a.sessions.put(token, adminUser{id: rows[0]["id"], username: rows[0]["username"], level: level, expires: expires})
	return map[string]any{"token": token, "username": rows[0]["username"], "level": level}, nil
}
