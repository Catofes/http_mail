package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GehirnInc/crypt/sha512_crypt"
)

type noStore struct{ t *testing.T }

func (s noStore) query(context.Context, string, ...any) ([]row, error) {
	s.t.Fatal("unexpected database query")
	return nil, nil
}
func (s noStore) execute(context.Context, string, ...any) error {
	s.t.Fatal("unexpected database write")
	return nil
}
func (s noStore) transaction(context.Context, ...statement) error {
	s.t.Fatal("unexpected database transaction")
	return nil
}

func perform(a *api, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	a.e.ServeHTTP(resp, req)
	return resp
}

func expectResponse(t *testing.T, got *httptest.ResponseRecorder, status int, body string) {
	t.Helper()
	if got.Code != status {
		t.Fatalf("status=%d body=%s, want status=%d", got.Code, got.Body.String(), status)
	}
	if body == "" {
		if got.Body.Len() != 0 {
			t.Fatalf("expected empty body, got %s", got.Body.String())
		}
		return
	}
	var actual, expected any
	if err := json.Unmarshal(got.Body.Bytes(), &actual); err != nil {
		t.Fatalf("invalid JSON response: %s", got.Body.String())
	}
	if err := json.Unmarshal([]byte(body), &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("body=%s, want %s", got.Body.String(), body)
	}
}

func expectError(t *testing.T, got *httptest.ResponseRecorder, code int) {
	t.Helper()
	e := apiError(code)
	expectResponse(t, got, e.status, fmt.Sprintf(`{"title":%q,"code":%d}`, e.Title, code))
}

func TestRequestCompatibility(t *testing.T) {
	for _, tt := range []struct {
		method, path, body string
		code               int
	}{
		{"GET", "/invite", "", 5}, {"GET", "/invite?code=", "", 5},
		{"GET", "/login", "", 21}, {"GET", "/login?token=", "", 21},
		{"GET", "/login?token=missing", "", 22}, {"DELETE", "/login?token=missing", "", 22},
		{"GET", "/server", "", 21}, {"GET", "/transport_default", "", 21},
		{"GET", "/domain", "", 21}, {"GET", "/domain/1", "", 21},
		{"POST", "/login", "", 16}, {"POST", "/login", "{}", 20},
		{"POST", "/login", `{"username":"abc"}`, 20},
		{"POST", "/login", `{"username":"abc","password":""}`, 20},
		{"POST", "/login", "{", 4}, {"POST", "/login", "{}{}", 4},
		{"GET", "/domain?token=missing", "{", 4},
		{"POST", "/register", "{", 4}, {"PUT", "/login", "", 5},
	} {
		t.Run(tt.method+tt.path+tt.body, func(t *testing.T) {
			a := newAPI(noStore{t}, config{SessionCacheSize: 2})
			expectError(t, perform(a, tt.method, tt.path, tt.body), tt.code)
		})
	}
}

func TestMediaTypesAndBodyCompatibility(t *testing.T) {
	a := newAPI(noStore{t}, config{SessionCacheSize: 2})
	for _, tt := range []struct {
		accept, contentType string
		status              int
		code                int
	}{
		{"application/json", "text/plain", 400, 2},
		{"application/*", "application/json; charset=utf-8", 400, 20},
		{"*/*", "application/json", 400, 20},
		{"text/html", "application/json", 400, -1},
		{"application/json;q=0, */*;q=1", "application/json", 400, -1},
		{"application/json", "", 500, -1},
	} {
		t.Run(tt.accept+tt.contentType, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/login", strings.NewReader("{}"))
			req.Header.Set("Accept", tt.accept)
			req.Header.Set("Content-Type", tt.contentType)
			resp := httptest.NewRecorder()
			a.e.ServeHTTP(resp, req)
			if tt.code >= 0 {
				expectError(t, resp, tt.code)
			} else if resp.Code != tt.status {
				t.Fatalf("status=%d, want %d", resp.Code, tt.status)
			}
		})
	}
	for _, body := range []string{"[]", "null", `{"username":123,"password":"abc"}`} {
		resp := perform(a, "POST", "/login", body)
		if resp.Code != 500 {
			t.Fatalf("non-object/non-string legacy input should return 500: %s", resp.Body.String())
		}
	}
	for _, tt := range []struct {
		length int64
		code   int
	}{{1, 3}, {-1, 16}} {
		req := httptest.NewRequest("POST", "/login", nil)
		req.Body = io.NopCloser(strings.NewReader(""))
		req.ContentLength = tt.length
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		a.e.ServeHTTP(resp, req)
		expectError(t, resp, tt.code)
	}
	req := httptest.NewRequest("GET", "/login", nil)
	req.Header.Set("Accept", "application/xml")
	resp := httptest.NewRecorder()
	a.e.ServeHTTP(resp, req)
	if resp.Code != 400 || resp.Body.String() != "<error><title>Json Required.</title><code>2</code></error>" {
		t.Fatalf("legacy XML negotiation changed: %d %s", resp.Code, resp.Body.String())
	}
}

func TestRoutingCompatibility(t *testing.T) {
	a := newAPI(noStore{t}, config{SessionCacheSize: 2})
	for _, path := range []string{"/", "/missing", "/user", "/dkim", "/transport_default/1", "/domain/1/2"} {
		t.Run(path, func(t *testing.T) {
			expectResponse(t, perform(a, "GET", path, ""), 404, "")
		})
	}
	for _, tt := range []struct{ path, allow string }{
		{"/invite", "GET, OPTIONS"}, {"/login", "DELETE, GET, OPTIONS, POST, PUT"},
		{"/register", "OPTIONS, POST"}, {"/server", "GET, OPTIONS"},
		{"/domain", "GET, OPTIONS, POST"}, {"/domain/1", "DELETE, GET, OPTIONS"},
		{"/user/1", "GET, OPTIONS, POST"}, {"/user/1/2", "DELETE, GET, OPTIONS, PUT"},
		{"/dkim/1", "GET, OPTIONS, PUT"}, {"/alias/1", "GET, OPTIONS, POST"},
		{"/alias/1/2", "DELETE, GET, OPTIONS"}, {"/bcc/1", "GET, OPTIONS, POST"},
		{"/bcc/1/2", "DELETE, GET, OPTIONS"}, {"/transport/1", "GET, OPTIONS, POST"},
		{"/transport/1/2", "DELETE, GET, OPTIONS"}, {"/transport_default", "GET, OPTIONS"},
		{"/transport_default/1/1", "OPTIONS, POST"},
	} {
		resp := perform(a, "OPTIONS", tt.path, "")
		expectResponse(t, resp, 200, "")
		if resp.Header().Get("Allow") != tt.allow {
			t.Fatalf("%s Allow=%s, want %s", tt.path, resp.Header().Get("Allow"), tt.allow)
		}
	}
	expectResponse(t, perform(a, "HEAD", "/domain", ""), 405, "")
	resp := perform(a, "PATCH", "/domain", "")
	expectResponse(t, resp, 405, "")
	if resp.Header().Get("Allow") != "GET, OPTIONS, POST" || resp.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("framework error headers changed: %v", resp.Header())
	}
	expectResponse(t, perform(a, "OPTIONS", "/domain/1/2", ""), 404, "")
	expectError(t, perform(a, "GET", "/domain/", ""), 21)
}

func TestPasswordsAndTokens(t *testing.T) {
	if got := adminPassword("321", "in.box.moe"); got != "96a1eae98a028a755a813e0da030c61eca30dd590dba0a78711cff70c48fa42d" {
		t.Fatalf("admin password incompatible: %s", got)
	}
	// Independently generated using OpenSSL passwd -6 (same format as Python crypt).
	const expected = "$6$0123456789abcdef$vNATSYYTivQfXwPTUT4q.sRFLs/sgxDXaPipzRlX3WOO4r1NcR.Og5OoU2Cd2agm1WA3pCJ30JU4EKMxpZaDy/"
	hash, err := sha512_crypt.New().Generate([]byte("test"), []byte("$6$0123456789abcdef"))
	if err != nil || hash != expected {
		t.Fatalf("SHA512-crypt incompatible: %s %v", hash, err)
	}
	for _, password := range []string{"test", "", "with spaces"} {
		hash, err := mailboxPassword(password)
		if err != nil || !regexp.MustCompile(`^\$6\$[a-f0-9]{16}\$[./a-zA-Z0-9]{86}$`).MatchString(hash) {
			t.Fatalf("bad mailbox hash: %s %v", hash, err)
		}
		if err := sha512_crypt.New().Verify(hash, []byte(password)); err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	for range 100 {
		token, err := generateToken()
		if err != nil || !regexp.MustCompile(`^[a-f0-9]{8,16}$`).MatchString(token) || seen[token] {
			t.Fatalf("bad token: %s %v", token, err)
		}
		seen[token] = true
	}
}

func TestSessions(t *testing.T) {
	s := newSessions(2)
	u := adminUser{id: 1, expires: time.Now().Add(time.Hour)}
	s.put("first", u)
	s.put("second", u)
	if got, ok := s.get("first"); !ok || got.expires != u.expires {
		t.Fatal("session lookup changed expiry")
	}
	s.put("third", u)
	if _, ok := s.get("second"); ok {
		t.Fatal("cache capacity not enforced")
	}
	s.remove("first")
	if _, ok := s.get("first"); ok {
		t.Fatal("logout did not remove session")
	}
	u.expires = time.Now().Add(-time.Second)
	s.put("expired", u)
	if _, ok := s.get("expired"); ok {
		t.Fatal("expired token still accepted")
	}
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Go(func() {
			token := fmt.Sprint(i)
			for range 100 {
				s.put(token, adminUser{expires: time.Now().Add(time.Hour)})
				s.get(token)
				s.remove(token)
			}
		})
	}
	wg.Wait()
}

func TestConfig(t *testing.T) {
	path := t.TempDir() + "/config.json"
	if err := os.WriteFile(path, []byte(`{"db_passwd":"a'b\\c","password_salt":"old-salt","db_maxshared":40}`), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := loadConfig(path)
	if err != nil || c.PasswordSalt != "old-salt" || !strings.Contains(c.dsn(), `password='a\'b\\c'`) {
		t.Fatalf("configuration incompatible: %v %s", err, c.dsn())
	}
}
