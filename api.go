package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"mime"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type api struct {
	db       store
	config   config
	sessions *sessions
	e        *echo.Echo
}

func newAPI(db store, c config) *api {
	a := &api{db: db, config: c, sessions: newSessions(c.SessionCacheSize), e: echo.New()}
	a.e.HideBanner = true
	a.e.HidePort = true
	// Echo's default error JSON differs from Falcon's public API.
	a.e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		var framework *echo.HTTPError
		if errors.As(err, &framework) && (framework.Code == 404 || framework.Code == 405) {
			if framework.Code == 405 {
				allowed := strings.Split(c.Response().Header().Get("Allow"), ", ")
				sort.Strings(allowed)
				c.Response().Header().Set("Allow", strings.Join(allowed, ", "))
			}
			_ = c.NoContent(framework.Code)
			return
		}
		writeError(c.Response(), c.Request(), err)
	}
	a.e.Pre(middleware.RemoveTrailingSlash(), a.translateJSON)
	a.e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{DisablePrintStack: true}))
	a.e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Echo's final path parameter can include slashes; Falcon used one segment.
			for _, name := range c.ParamNames() {
				if value := c.Param(name); value == "" || strings.Contains(value, "/") {
					return echo.ErrNotFound
				}
			}
			return next(c)
		}
	})
	for _, route := range []struct {
		path    string
		methods []string
	}{
		{"/invite", []string{"GET"}},
		{"/login", []string{"GET", "POST", "PUT", "DELETE"}},
		{"/register", []string{"POST"}},
		{"/server", []string{"GET"}},
		{"/domain", []string{"GET", "POST"}},
		{"/domain/:domain_id", []string{"GET", "DELETE"}},
		{"/user/:domain_id", []string{"GET", "POST"}},
		{"/user/:domain_id/:user_id", []string{"GET", "PUT", "DELETE"}},
		{"/dkim/:domain_id", []string{"GET", "PUT"}},
		{"/alias/:domain_id", []string{"GET", "POST"}},
		{"/alias/:domain_id/:alias_id", []string{"GET", "DELETE"}},
		{"/bcc/:domain_id", []string{"GET", "POST"}},
		{"/bcc/:domain_id/:bcc_id", []string{"GET", "DELETE"}},
		{"/transport/:domain_id", []string{"GET", "POST"}},
		{"/transport/:domain_id/:transport_id", []string{"GET", "DELETE"}},
		{"/transport_default", []string{"GET"}},
		{"/transport_default/:domain_id/:operate_id", []string{"POST"}},
	} {
		for _, method := range route.methods {
			a.e.Add(method, route.path, func(c echo.Context) error {
				r := c.Get("legacyRequest").(*request)
				result, err := a.dispatch(r)
				if err != nil {
					return err
				}
				writeJSON(c.Response(), 200, result)
				return nil
			})
		}
		allowed := append(append([]string{}, route.methods...), "OPTIONS")
		sort.Strings(allowed)
		a.e.OPTIONS(route.path, func(c echo.Context) error {
			c.Response().Header().Set("Allow", strings.Join(allowed, ", "))
			return c.NoContent(200)
		})
	}
	return a
}

type request struct {
	*http.Request
	parts   []string
	body    any
	hasBody bool
	user    adminUser
}

func (r *request) object() (map[string]any, error) {
	if !r.hasBody {
		return nil, apiError(16)
	}
	obj, ok := r.body.(map[string]any)
	if !ok {
		return nil, errors.New("request body is not an object")
	}
	return obj, nil
}

func stringField(obj map[string]any, name string, missingCode int) (string, error) {
	v, ok := obj[name]
	if !ok {
		return "", apiError(missingCode)
	}
	s, ok := v.(string)
	if !ok {
		return "", errors.New("request field is not a string")
	}
	return s, nil
}

func (a *api) translateJSON(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		c.Response().Header().Set("Content-Type", "application/json")
		r := &request{Request: req, parts: strings.Split(strings.TrimSuffix(req.URL.Path, "/"), "/")[1:]}
		if !accepts(req.Header.Get("Accept"), "application/json") {
			return apiError(2)
		}
		if req.Method == "POST" || req.Method == "PUT" {
			contentType := req.Header.Get("Content-Type")
			if contentType == "" {
				// The old middleware raised TypeError for a missing Content-Type.
				return errors.New("missing content type")
			}
			if !strings.Contains(contentType, "application/json") {
				return apiError(2)
			}
		}
		// WSGI did not read bodies without a positive Content-Length.
		if req.ContentLength > 0 {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return errors.New("could not read request")
			}
			if len(body) == 0 {
				return apiError(3)
			}
			if !utf8.Valid(body) || json.Unmarshal(body, &r.body) != nil {
				return apiError(4)
			}
			r.hasBody = true
		}
		c.Set("legacyRequest", r)
		return next(c)
	}
}

func (a *api) dispatch(r *request) (any, error) {
	name := r.parts[0]
	if name == "invite" || name == "register" || name == "login" && r.Method == "PUT" {
		used := 0
		if name == "login" {
			used = 1
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			return nil, apiError(5)
		}
		rows, err := a.db.query(r.Context(), "SELECT * FROM invite_codes WHERE code = $1 AND used = $2", code, used)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, apiError(6)
		}
		if name == "invite" {
			return nil, nil
		}
		return a.account(r, code)
	}
	if name == "login" && r.Method == "POST" {
		return a.account(r, "")
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		return nil, apiError(21)
	}
	u, ok := a.sessions.get(token)
	if !ok {
		return nil, apiError(22)
	}
	r.user = u
	if len(r.parts) >= 2 && u.level != 100 {
		rows, err := a.db.query(r.Context(), "SELECT * FROM virtual_domains WHERE id = $1 AND admin_user_id = $2", r.parts[1], u.id)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, apiError(24)
		}
	}
	switch name {
	case "login":
		if r.Method == "DELETE" {
			a.sessions.remove(token)
		}
		return nil, nil
	case "server":
		return a.list(r, "SELECT domain_name, server_mark, region_mark, default_mark FROM my_networks")
	case "domain":
		return a.domain(r)
	case "user":
		return a.mailbox(r)
	case "dkim":
		return a.dkim(r)
	case "alias", "bcc", "transport":
		return a.mailRule(r)
	case "transport_default":
		return a.defaultTransport(r)
	}
	return nil, apiError(0)
}

func (a *api) list(r *request, query string, args ...any) (any, error) {
	rows, err := a.db.query(r.Context(), query, args...)
	return map[string]any{"result": rows}, err
}

func (a *api) one(r *request, code int, query string, args ...any) (any, error) {
	rows, err := a.db.query(r.Context(), query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apiError(code)
	}
	return map[string]any{"result": rows[0]}, nil
}

// Falcon's client_accepts_json honors wildcard media ranges and q-values.
func accepts(header, target string) bool {
	if header == "" {
		return true
	}
	best, quality := -1, 0.0
	for _, item := range strings.Split(header, ",") {
		media, params, err := mime.ParseMediaType(strings.TrimSpace(item))
		if err != nil {
			continue
		}
		specificity := -1
		switch media {
		case target:
			specificity = 2
		case strings.Split(target, "/")[0] + "/*":
			specificity = 1
		case "*/*":
			specificity = 0
		}
		q := 1.0
		if value, exists := params["q"]; exists {
			parsed, err := strconv.ParseFloat(value, 64)
			if err == nil && parsed >= 0 && parsed <= 1 {
				q = parsed
			}
		}
		if specificity > best || specificity == best && q > quality {
			best, quality = specificity, q
		}
	}
	return quality > 0 && best >= 0
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	var data []byte
	if value != nil {
		var err error
		data, err = json.Marshal(value)
		if err != nil {
			status, data = 500, []byte(`{"title":"Unknown Error.","code":0}`)
		}
	}
	w.WriteHeader(status)
	if len(data) != 0 {
		_, _ = w.Write(data)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var legacy *legacyError
	if !errors.As(err, &legacy) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(500)
		_, _ = io.WriteString(w, "A server error occurred.  Please contact the administrator.")
		return
	}
	if accepts(r.Header.Get("Accept"), "application/json") {
		writeJSON(w, legacy.status, legacy)
	} else if accepts(r.Header.Get("Accept"), "application/xml") || accepts(r.Header.Get("Accept"), "text/xml") {
		body, _ := xml.Marshal(struct {
			XMLName xml.Name `xml:"error"`
			Title   string   `xml:"title"`
			Code    int      `xml:"code"`
		}{Title: legacy.Title, Code: legacy.Code})
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(legacy.status)
		_, _ = w.Write(body)
	} else {
		w.WriteHeader(legacy.status)
	}
}
