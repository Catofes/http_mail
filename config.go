package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Names match the settings in the old RConfig; no database migration is needed.
type config struct {
	PasswordSalt     string `json:"password_salt"`
	DBUser           string `json:"db_user"`
	DBPassword       string `json:"db_passwd"`
	DBHost           string `json:"db_host"`
	DBPort           int    `json:"db_port"`
	DBName           string `json:"db_db"`
	DBMinCached      int    `json:"db_mincached"`
	DBMaxConnections int    `json:"db_maxconnections"`
	SessionCacheSize int    `json:"session_cache_size"`
	DBSSLMode        string `json:"db_sslmode"`
}

func loadConfig(path string) (config, error) {
	c := config{PasswordSalt: "in.box.moe", DBUser: "mailserver", DBHost: "127.0.0.1", DBPort: 5432,
		DBName: "mailserver", DBMinCached: 5, DBMaxConnections: 40, SessionCacheSize: 1000, DBSSLMode: "prefer"}
	data, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	if err = json.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("invalid config JSON: %w", err)
	}
	if c.DBPort < 1 || c.DBPort > 65535 || c.DBMinCached < 0 || c.DBMaxConnections < 1 ||
		c.DBMinCached > c.DBMaxConnections || c.SessionCacheSize < 1 {
		return c, fmt.Errorf("invalid database pool, port or session cache settings")
	}
	return c, nil
}

func (c config) dsn() string {
	quote := func(s string) string {
		return "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(s) + "'"
	}
	return "host=" + quote(c.DBHost) + " port=" + strconv.Itoa(c.DBPort) + " user=" + quote(c.DBUser) +
		" password=" + quote(c.DBPassword) + " dbname=" + quote(c.DBName) + " sslmode=" + quote(c.DBSSLMode) + " connect_timeout=3"
}
