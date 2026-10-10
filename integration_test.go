package main

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/GehirnInc/crypt/sha512_crypt"
	"github.com/lib/pq"
)

func testPostgres(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("HTTP_MAIL_TEST_DSN")
	if dsn == "" {
		t.Skip("run make integration to use a disposable PostgreSQL database")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		if err != nil {
			t.Fatal(err)
		}
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Close() })
	token, err := generateToken()
	if err != nil {
		t.Fatal(err)
	}
	schema := "http_mail_test_" + token
	if _, err := admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE"); err != nil {
			t.Error(err)
		}
	})
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	fixture, err := os.ReadFile("testdata/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(fixture)); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPostgresCompatibility(t *testing.T) {
	db := testPostgres(t)
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec("INSERT INTO invite_codes(code,used) VALUES ('new',0),('used',1),('reset',1),('duplicate',0)")
	for _, user := range []struct {
		name  string
		level int
	}{{"root", 100}, {"owner", 1}, {"other", 1}, {"power", 5}, {"notroot", 101}} {
		exec("INSERT INTO admin_users(username,password,level) VALUES ($1,$2,$3)", user.name, adminPassword("321", "in.box.moe"), user.level)
	}
	exec("INSERT INTO my_networks(domain_name,server_mark,region_mark,default_mark) VALUES ('mx.example.com','SFDO','3CN','0default')")
	a := newAPI(&postgresStore{db}, config{PasswordSalt: "in.box.moe", SessionCacheSize: 1000})
	call := func(method, path, token, body string) *httptest.ResponseRecorder {
		t.Helper()
		if token != "" {
			separator := "?"
			if strings.Contains(path, "?") {
				separator = "&"
			}
			path += separator + "token=" + token
		}
		return perform(a, method, path, body)
	}
	login := func(name string) string {
		t.Helper()
		resp := call("POST", "/login", "", fmt.Sprintf(`{"username":%q,"password":"321"}`, name))
		if resp.Code != 200 {
			t.Fatalf("login: %d %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Token    string
			Username string
			Level    int
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Token == "" || body.Username != name {
			t.Fatalf("invalid login: %s", resp.Body.String())
		}
		return body.Token
	}
	root, owner, other, power, notroot := login("root"), login("owner"), login("other"), login("power"), login("notroot")
	expectResponse(t, call("GET", "/login", owner, ""), 200, "")
	expectResponse(t, call("GET", "/invite?code=new", "", ""), 200, "")
	expectError(t, call("GET", "/invite?code=used", "", ""), 6)
	expectError(t, call("POST", "/register?code=missing", "", `{}`), 6)
	expectError(t, call("POST", "/register?code=new", "", `{"username":"a!","password":"321"}`), 13)
	expectError(t, call("POST", "/register?code=duplicate", "", `{"username":"owner","password":"321"}`), 19)
	expectResponse(t, call("POST", "/register?code=new", "", `{"username":"newbie","password":"321","ignored":true}`), 200, "")
	expectError(t, call("GET", "/invite?code=new", "", ""), 6)
	newbie := login("newbie")
	expectError(t, call("POST", "/login", "", `{"username":"newbie","password":"bad"}`), 25)
	expectError(t, call("PUT", "/login?code=duplicate", "", `{"password":"new"}`), 6)
	expectResponse(t, call("PUT", "/login?code=new", "", `{"password":"changed"}`), 200, "")
	expectResponse(t, call("GET", "/login", newbie, ""), 200, "") // reset did not revoke cached sessions
	expectError(t, call("POST", "/login", "", `{"username":"newbie","password":"321"}`), 25)
	var stored string
	if err := db.QueryRow("SELECT password FROM admin_users WHERE username='newbie'").Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored != adminPassword("changed", "in.box.moe") {
		t.Fatal("registration/reset password format changed")
	}
	expectResponse(t, call("DELETE", "/login", newbie, ""), 200, "")
	expectError(t, call("GET", "/login", newbie, ""), 22)

	expectResponse(t, call("GET", "/domain", owner, ""), 200, `{"result":null}`)
	expectError(t, call("POST", "/domain", owner, `{}`), 7)
	expectError(t, call("POST", "/domain", owner, `{"domain":"invalid"}`), 9)
	expectResponse(t, call("POST", "/domain", owner, `{"domain":"example.com"}`), 200, `{"result":{"id":1,"name":"example.com"}}`)
	expectError(t, call("POST", "/domain", other, `{"domain":"example.com"}`), 8)
	expectResponse(t, call("POST", "/domain", other, `{"domain":"other.example.com"}`), 200, `{"result":{"id":2,"name":"other.example.com"}}`)
	expectResponse(t, call("GET", "/domain", owner, ""), 200, `{"result":[{"id":1,"name":"example.com"}]}`)
	expectResponse(t, call("GET", "/domain/1", owner, ""), 200, `{"result":{"id":1,"name":"example.com"}}`)
	expectError(t, call("GET", "/domain/1", other, ""), 24)
	expectError(t, call("GET", "/domain/1", notroot, ""), 24) // only level == 100 bypasses ownership
	expectResponse(t, call("GET", "/domain/1", root, ""), 200, `{"result":{"id":1,"name":"example.com"}}`)
	expectResponse(t, call("GET", "/domain", root, ""), 200, `{"result":[{"id":1,"name":"example.com"},{"id":2,"name":"other.example.com"}]}`)
	expectError(t, call("GET", "/domain/999", root, ""), 24)
	expectError(t, call("GET", "/domain/not-a-number", root, ""), 1)
	expectResponse(t, call("GET", "/server", owner, ""), 200, `{"result":[{"domain_name":"mx.example.com","server_mark":"SFDO","region_mark":"3CN","default_mark":"0default"}]}`)

	expectResponse(t, call("GET", "/user/1", owner, ""), 200, `{"result":null}`)
	expectError(t, call("GET", "/user/1", other, ""), 24)
	expectError(t, call("POST", "/user/1", owner, `{"username":"!"}`), 13)
	expectError(t, call("POST", "/user/1", owner, `{"username":"box"}`), 14)
	expectResponse(t, call("POST", "/user/1", owner, `{"username":"box","password":"mail-password"}`), 200, "")
	expectError(t, call("POST", "/user/1", owner, `{"username":"box","password":"new"}`), 17)
	expectResponse(t, call("GET", "/user/1", owner, ""), 200, `{"result":[{"id":1,"domain_id":1,"email":"box@example.com"}]}`)
	expectResponse(t, call("GET", "/user/1/1", owner, ""), 200, `{"result":{"id":1,"domain_id":1,"email":"box@example.com"}}`)
	expectError(t, call("GET", "/user/2/1", other, ""), 26)
	if err := db.QueryRow("SELECT password FROM virtual_users WHERE id=1").Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if err := sha512_crypt.New().Verify(stored, []byte("mail-password")); err != nil {
		t.Fatal(err)
	}
	expectError(t, call("PUT", "/user/1/1", owner, `{}`), 14)
	expectResponse(t, call("PUT", "/user/1/1", owner, `{"password":"changed"}`), 200, "")
	if err := db.QueryRow("SELECT password FROM virtual_users WHERE id=1").Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if err := sha512_crypt.New().Verify(stored, []byte("changed")); err != nil {
		t.Fatal(err)
	}
	expectResponse(t, call("PUT", "/user/1/999", owner, `{"password":"unused"}`), 200, "")
	expectResponse(t, call("DELETE", "/user/1/999", owner, ""), 200, "")

	for _, rule := range []struct {
		path, table, source, storedSource string
		code                              int
	}{
		{"alias", "virtual_aliases", "", "@example.com", 29},
		{"bcc", "recipient_bcc", "", "@example.com", 32},
		{"transport", "transport_domains", "box@", "box@example.com", 33},
	} {
		t.Run(rule.path, func(t *testing.T) {
			path := "/" + rule.path + "/1"
			expectResponse(t, call("GET", path, owner, ""), 200, `{"result":null}`)
			expectError(t, call("POST", path, owner, `{}`), 27)
			expectError(t, call("POST", path, power, `{}`), 24) // ownership precedes level/body checks
			expectError(t, call("POST", path, root, `{}`), 20)
			body := fmt.Sprintf(`{"source":%q,"destination":"target@sub.example.net","region":"SFDO"}`, rule.source)
			expectResponse(t, call("POST", path, root, body), 200, "")
			var source string
			if err := db.QueryRow("SELECT source FROM " + rule.table + " WHERE id=1").Scan(&source); err != nil {
				t.Fatal(err)
			}
			if source != rule.storedSource {
				t.Fatalf("source=%q, want %q", source, rule.storedSource)
			}
			expected := fmt.Sprintf(`{"id":1,"source":%q,"destination":"target@sub.example.net"`, rule.storedSource)
			if rule.path != "alias" {
				expected += `,"region":"SFDO"`
			}
			expected += `}`
			expectResponse(t, call("GET", path, owner, ""), 200, `{"result":[`+expected+`]}`)
			expectResponse(t, call("GET", path+"/1", owner, ""), 200, `{"result":`+expected+`}`)
			expectError(t, call("GET", path+"/999", owner, ""), rule.code)
			expectError(t, call("DELETE", path+"/1", owner, ""), 27)
			expectResponse(t, call("DELETE", path+"/1", root, ""), 200, "")
			expectResponse(t, call("DELETE", path+"/999", root, ""), 200, "")
			expectResponse(t, call("GET", path, owner, ""), 200, `{"result":null}`)
		})
	}
	for _, path := range []string{"alias", "bcc"} {
		expectError(t, call("POST", "/"+path+"/1", root, `{"source":"!","destination":"target@example.com","region":"SFDO"}`), 13)
		expectError(t, call("POST", "/"+path+"/1", root, `{"source":"valid","destination":"invalid","region":"SFDO"}`), 28)
	}
	expectError(t, call("POST", "/transport/1", root, `{"source":"box","destination":"smtp:any","region":"SFDO"}`), 13)
	expectResponse(t, call("POST", "/domain", power, `{"domain":"power.example.com"}`), 200, `{"result":{"id":3,"name":"power.example.com"}}`)
	expectResponse(t, call("POST", "/alias/3", power, `{"source":"","destination":"target@example.com"}`), 200, "")

	expectResponse(t, call("GET", "/dkim/1", owner, ""), 200, "")
	expectError(t, call("PUT", "/dkim/1", owner, `{}`), 20)
	expectResponse(t, call("PUT", "/dkim/1", owner, `{"selector":"old","private_key":"old-key"}`), 200, "")
	oldKey := sha512.Sum512([]byte("old-key"))
	expectResponse(t, call("GET", "/dkim/1", owner, ""), 200, fmt.Sprintf(`{"domain":"example.com","selector":"old","key_sha512":%q}`, hex.EncodeToString(oldKey[:])))
	expectResponse(t, call("PUT", "/dkim/1", owner, `{"selector":"new","private_key":"new-key"}`), 200, "")
	newKey := sha512.Sum512([]byte("new-key"))
	expectResponse(t, call("GET", "/dkim/1", owner, ""), 200, fmt.Sprintf(`{"domain":"example.com","selector":"new","key_sha512":%q}`, hex.EncodeToString(newKey[:])))
	var count int
	if err := db.QueryRow("SELECT count(*) FROM opendkim_keys").Scan(&count); err != nil || count != 1 {
		t.Fatalf("DKIM replaced incorrectly: %d %v", count, err)
	}
	exec("INSERT INTO opendkim_signings(author,dkim_id) VALUES ('other.example.com',NULL)")
	expectResponse(t, call("GET", "/dkim/2", other, ""), 200, "")

	expectResponse(t, call("GET", "/transport_default", owner, ""), 200, `{"result":{"1":{"Illustrate":"Relay all mail to one server and store in it. This is helpful when you want to add multi servers in your mx record.","parameters":"You need a json context with you server id in it. Like {\"id\":1}"}}}`)
	expectError(t, call("POST", "/transport_default/1/2", owner, `{}`), 31)
	expectError(t, call("POST", "/transport_default/1/1", owner, `{}`), 20)
	expectError(t, call("POST", "/transport_default/1/1", owner, `{"id":999}`), 30)
	expectResponse(t, call("POST", "/transport/1", root, `{"source":"","destination":"smtp:old","region":"old"}`), 200, "")
	expectResponse(t, call("POST", "/transport_default/1/1", owner, `{"id":1}`), 200, "") // no level-5 requirement
	var destination, region string
	if err := db.QueryRow("SELECT destination,region FROM transport_domains WHERE domain_id=1").Scan(&destination, &region); err != nil {
		t.Fatal(err)
	}
	if destination != "lmtp:unix:private/dovecot-lmtp" || region != "0default" {
		t.Fatalf("legacy default transport changed: %s %s", destination, region)
	}
	exec("ALTER TABLE transport_domains ADD CONSTRAINT reject_default CHECK (region <> '0default') NOT VALID")
	expectError(t, call("POST", "/transport_default/1/1", owner, `{"id":1}`), 1)
	if err := db.QueryRow("SELECT count(*) FROM transport_domains WHERE domain_id=1 AND region='0default'").Scan(&count); err != nil || count != 1 {
		t.Fatalf("transaction failed to roll back: %d %v", count, err)
	}
	exec("ALTER TABLE transport_domains DROP CONSTRAINT reject_default")
	expectResponse(t, call("POST", "/transport_default/1/1", owner, `{"id":1}`), 200, "")

	expectResponse(t, call("DELETE", "/user/1/1", owner, ""), 200, "")
	expectError(t, call("GET", "/user/1/1", owner, ""), 26)
	expectResponse(t, call("POST", "/user/1", owner, `{"username":"remaining","password":"321"}`), 200, "")
	expectResponse(t, call("DELETE", "/domain/1", owner, ""), 200, "")
	if err := db.QueryRow("SELECT count(*) FROM virtual_users WHERE domain_id=1").Scan(&count); err != nil || count != 0 {
		t.Fatalf("domain deletion left mailbox users: %d %v", count, err)
	}
	expectError(t, call("GET", "/domain/1", owner, ""), 24)
	expectResponse(t, call("DELETE", "/domain/999", root, ""), 200, "")

	// An expired session remains invalid; requests never renew the 24h deadline.
	a.sessions.put("expired", adminUser{id: 2, level: 1, expires: time.Now().Add(-time.Second)})
	expectError(t, call("GET", "/login", "expired", ""), 22)
	rows, err := (&postgresStore{db}).query(context.Background(), "SELECT id FROM virtual_users WHERE domain_id=$1", 1)
	if err != nil || rows != nil {
		t.Fatalf("empty query must produce null: %v %v", rows, err)
	}
}
