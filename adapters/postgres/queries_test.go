package postgres

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/user"
	"testing"

	"github.com/nuveo/prest/config"
)

func TestMain(m *testing.M) {
	os.Setenv("PREST_CONF", "../../testdata/prest.toml")
	config.Load()
	createMockScripts(config.PrestConf.QueriesPath)
	writeMockScripts(config.PrestConf.QueriesPath)

	code := m.Run()

	removeMockScripts(config.PrestConf.QueriesPath)
	os.Exit(code)
}

func TestValidGetScript(t *testing.T) {
	var testCases = []struct {
		description string
		method      string
		path        string
		file        string
		err         error
	}{
		{"Get script file by GET Method", "GET", "fulltable", "get_all", nil},
		{"Get script file by POST Method", "POST", "fulltable", "write_all", nil},
		{"Get script file by PUT Method", "PUT", "fulltable", "put_all", nil},
		{"Get script file by PATCH Method", "PATCH", "fulltable", "patch_all", nil},
		{"Get script file by DELETE Method", "DELETE", "fulltable", "delete_all", nil},
	}

	for _, tc := range testCases {
		t.Log(tc.description)
		_, err := GetScript(tc.method, tc.path, tc.file)
		if err != tc.err {
			t.Errorf("expected no errors, but got %s", err)
		}
	}
}

func TestInvalidGetScript(t *testing.T) {
	var testCases = []struct {
		description string
		method      string
		path        string
		file        string
	}{
		{"Try get a script with INVALID HTTP Method", "ANY", "fulltable", "delete_all"},
		{"Try get a script noexistent", "GET", "fulltable", "dloohot_all"},
		{"Try get a script with nonexistent folder", "GET", "sasalla", "get_all"},
	}

	for _, tc := range testCases {
		t.Log(tc.description)
		_, err := GetScript(tc.method, tc.path, tc.file)
		if err == nil {
			t.Errorf("expected no error, but got %s", err)
		}
	}
}

func TestParseScript(t *testing.T) {
	queryURL := url.Values{}
	queryURL.Set("field1", "abc")

	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	scriptPath := fmt.Sprint(user.HomeDir, "/queries/fulltable/%s")

	t.Log("Parse script with get_all file")
	sql, _, err := ParseScript(fmt.Sprintf(scriptPath, "get_all.read.sql"), queryURL)
	if err != nil {
		t.Errorf("expected no error, but got: %v", err)
	}

	if sql != "SELECT * FROM test7 WHERE name = 'abc'" {
		t.Errorf("SQL unexpected, got: %s", sql)
	}

	// Add new values
	queryURL.Del("field1")
	queryURL.Set("notable", "123")

	t.Log("Try Parse Script with noexistent script")
	sql, _, err = ParseScript(fmt.Sprintf(scriptPath, "gt_all.read.sql"), queryURL)
	if err == nil {
		t.Errorf("expected error, but got: %v", err)
	}

	if sql != "" {
		t.Errorf("expected empty string, got: %s", sql)
	}
}

func TestValidWriteSQL(t *testing.T) {
	var testValidCases = []struct {
		description string
		sql         string
		values      []interface{}
		err         error
	}{
		{"Execute a valid INSERT sql", "INSERT INTO test7(name) values ('lulu')", []interface{}{}, nil},
		{"Execute a valid UPDATE sql", "UPDATE test7 SET name = 'lulu' WHERE surname = 'temer'", []interface{}{}, nil},
		{"Execute a valid DELETE sql", "DELETE FROM test7 WHERE name = 'lulu'", []interface{}{}, nil},
	}
	for _, tc := range testValidCases {
		t.Log(tc.description)
		_, err := WriteSQL(tc.sql, tc.values)
		if tc.err != err {
			t.Error(tc.err, err)
		}
	}
}

func TestInvalidWriteSQL(t *testing.T) {
	var testInvalidCases = []struct {
		description string
		sql         string
		values      []interface{}
	}{
		{"Execute an invalid INSERT sql", "INSERT INTO test7 (tool) values (lulu)", []interface{}{}},
		{"Execute an invalid UPDATE sql", "UPDATE test7 SET name = lulu WHERE surname =", []interface{}{}},
		{"Execute an invalid DELETE sql", "DELETE FROM test7 WHERE name = lulu AND surname =", []interface{}{}},
	}

	for _, tc := range testInvalidCases {
		t.Log(tc.description)
		_, err := WriteSQL(tc.sql, tc.values)
		if err == nil {
			t.Errorf("expected nil, but got %v", err)
		}
	}
}

func TestExecuteScripts(t *testing.T) {
	var testCases = []struct {
		description string
		method      string
		sql         string
		values      []interface{}
		err         error
	}{
		{"Get result with GET HTTP Method", "GET", "SELECT * FROM test7", []interface{}{}, nil},
		{"Get result with POST HTTP Method", "POST", "INSERT INTO test7 (name) VALUES ('lala')", []interface{}{}, nil},
		{"Get result with PUT HTTP Method", "PUT", "UPDATE test7 SET name = 'lala' WHERE surname = 'temer'", []interface{}{}, nil},
		{"Get result with PATCH HTTP Method", "PATCH", "UPDATE test7 SET surname = 'temer' WHERE name = 'lala'", []interface{}{}, nil},
		{"Get result with DELETE HTTP Method", "DELETE", "DELETE FROM test7 WHERE surname = 'lala'", []interface{}{}, nil},
	}

	for _, tc := range testCases {
		t.Log(tc.description)
		_, err := ExecuteScripts(tc.method, tc.sql, tc.values)
		if tc.err != err {
			t.Errorf("expected no errors, but got %s", err)
		}
	}

	t.Log("Get errors with invalid HTTP Method")
	values := make([]interface{}, 0)
	result, err := ExecuteScripts("ANY", "SELECT * FROM test7", values)
	if len(result) > 0 {
		t.Errorf("expected empty result, but got %s", result)
	}

	if err == nil {
		t.Errorf("expected errors, but got %s", err)
	}
}
