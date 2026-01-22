package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	log "github.com/gophish/gophish/logger"
)

var validConfig = []byte(`{
	"admin_server": {
		"listen_url": "127.0.0.1:3333",
		"use_tls": true,
		"cert_path": "gophish_admin.crt",
		"key_path": "gophish_admin.key"
	},
	"phish_server": {
		"listen_url": "0.0.0.0:8080",
		"use_tls": false,
		"cert_path": "example.crt",
		"key_path": "example.key"
	},
	"db_name": "sqlite3",
	"db_path": "gophish.db",
	"migrations_prefix": "db/db_",
	"contact_address": ""
}`)

func createTemporaryConfig(t *testing.T) *os.File {
	f, err := ioutil.TempFile("", "gophish-config")
	if err != nil {
		t.Fatalf("unable to create temporary config: %v", err)
	}
	return f
}

func removeTemporaryConfig(t *testing.T, f *os.File) {
	err := f.Close()
	if err != nil {
		t.Fatalf("unable to remove temporary config: %v", err)
	}
}

func TestLoadConfig(t *testing.T) {
	f := createTemporaryConfig(t)
	defer removeTemporaryConfig(t, f)
	_, err := f.Write(validConfig)
	if err != nil {
		t.Fatalf("error writing config to temporary file: %v", err)
	}
	// Load the valid config
	conf, err := LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("error loading config from temporary file: %v", err)
	}

	expectedConfig := &Config{}
	err = json.Unmarshal(validConfig, &expectedConfig)
	if err != nil {
		t.Fatalf("error unmarshaling config: %v", err)
	}
	expectedConfig.MigrationsPath = expectedConfig.MigrationsPath + expectedConfig.DBName
	expectedConfig.TestFlag = false
	expectedConfig.AdminConf.CSRFKey = ""
	expectedConfig.Logging = &log.Config{}
	expectedConfig.ServerName = "Apache/2.4.41 (Ubuntu)"
	expectedConfig.SessionCookieName = "PHPSESSID"
	if !reflect.DeepEqual(expectedConfig, conf) {
		t.Fatalf("invalid config received. expected %#v got %#v", expectedConfig, conf)
	}

	// Load an invalid config
	_, err = LoadConfig("bogusfile")
	if err == nil {
		t.Fatalf("expected error when loading invalid config, but got %v", err)
	}
}

var customOpsecConfig = []byte(`{
	"admin_server": {
		"listen_url": "127.0.0.1:3333",
		"use_tls": true,
		"cert_path": "gophish_admin.crt",
		"key_path": "gophish_admin.key"
	},
	"phish_server": {
		"listen_url": "0.0.0.0:8080",
		"use_tls": false,
		"cert_path": "example.crt",
		"key_path": "example.key"
	},
	"db_name": "sqlite3",
	"db_path": "gophish.db",
	"migrations_prefix": "db/db_",
	"server_name": "nginx/1.18.0",
	"session_cookie_name": "JSESSIONID"
}`)

func TestLoadConfigCustomOpsec(t *testing.T) {
	f := createTemporaryConfig(t)
	defer removeTemporaryConfig(t, f)
	_, err := f.Write(customOpsecConfig)
	if err != nil {
		t.Fatalf("error writing config to temporary file: %v", err)
	}

	conf, err := LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("error loading config from temporary file: %v", err)
	}

	if conf.ServerName != "nginx/1.18.0" {
		t.Errorf("ServerName = %s, want nginx/1.18.0", conf.ServerName)
	}
	if conf.SessionCookieName != "JSESSIONID" {
		t.Errorf("SessionCookieName = %s, want JSESSIONID", conf.SessionCookieName)
	}
}

func TestLoadConfigDefaultOpsec(t *testing.T) {
	f := createTemporaryConfig(t)
	defer removeTemporaryConfig(t, f)
	_, err := f.Write(validConfig)
	if err != nil {
		t.Fatalf("error writing config to temporary file: %v", err)
	}

	conf, err := LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("error loading config from temporary file: %v", err)
	}

	if conf.ServerName != "Apache/2.4.41 (Ubuntu)" {
		t.Errorf("Default ServerName = %s, want Apache/2.4.41 (Ubuntu)", conf.ServerName)
	}
	if conf.SessionCookieName != "PHPSESSID" {
		t.Errorf("Default SessionCookieName = %s, want PHPSESSID", conf.SessionCookieName)
	}
}

func TestLoadConfigEncryptionKey(t *testing.T) {
	configWithKey := []byte(`{
		"admin_server": {"listen_url": "127.0.0.1:3333"},
		"phish_server": {"listen_url": "0.0.0.0:8080"},
		"db_name": "sqlite3",
		"db_path": "gophish.db",
		"migrations_prefix": "db/db_",
		"encryption_key": "0123456789abcdef0123456789abcdef"
	}`)

	f := createTemporaryConfig(t)
	defer removeTemporaryConfig(t, f)
	_, err := f.Write(configWithKey)
	if err != nil {
		t.Fatalf("error writing config: %v", err)
	}

	conf, err := LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("error loading config: %v", err)
	}

	if conf.EncryptionKey != "0123456789abcdef0123456789abcdef" {
		t.Errorf("EncryptionKey not loaded correctly")
	}
}
