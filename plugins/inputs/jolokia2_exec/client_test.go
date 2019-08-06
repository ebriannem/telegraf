package jolokia2_exec

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"github.com/influxdata/telegraf"
)

func TestJolokia2_ClientAuthRequest(t *testing.T) {
	var username string
	var password string
	var requests []map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, _ = r.BasicAuth()

		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &requests)
		if err != nil {
			t.Error(err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plugin := setupPlugin(t, fmt.Sprintf(`
		[jolokia2_exec_agent]
			urls = ["%s/jolokia"]
			username = "sally"
			password = "seashore"
		[[jolokia2_exec_agent.metric]]
			name  = "hello"
			mbean = "hello:foo=bar"
	`, server.URL))

	var acc testutil.Accumulator
	plugin.Gather(&acc)

	if username != "sally" {
		t.Errorf("Expected to post with username %s, but was %s", "sally", username)
	}
	if password != "seashore" {
		t.Errorf("Expected to post with password %s, but was %s", "seashore", password)
	}
	if len(requests) == 0 {
		t.Fatal("Expected to post a request body, but was empty.")
	}

	request := requests[0]
	if expect := "hello:foo=bar"; request["mbean"] != expect {
		t.Errorf("Expected to query mbean %s, but was %s", expect, request["mbean"])
	}
}

func TestJolokia2_ClientProxyAuthRequest(t *testing.T) {
	var requests []map[string]interface{}

	var username string
	var password string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, _ = r.BasicAuth()

		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &requests)
		if err != nil {
			t.Error(err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plugin := setupPlugin(t, fmt.Sprintf(`
		[jolokia2_exec_proxy]
			url = "%s/jolokia"
			username = "sally"
			password = "seashore"

		[[jolokia2_exec_proxy.target]]
			url = "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"
			username = "jack"
			password = "benimble"

		[[jolokia2_exec_proxy.metric]]
			name  = "hello"
			mbean = "hello:foo=bar"
	`, server.URL))

	var acc testutil.Accumulator
	plugin.Gather(&acc)

	if username != "sally" {
		t.Errorf("Expected to post with username %s, but was %s", "sally", username)
	}
	if password != "seashore" {
		t.Errorf("Expected to post with password %s, but was %s", "seashore", password)
	}
	if len(requests) == 0 {
		t.Fatal("Expected to post a request body, but was empty.")
	}

	request := requests[0]
	if expect := "hello:foo=bar"; request["mbean"] != expect {
		t.Errorf("Expected to query mbean %s, but was %s", expect, request["mbean"])
	}

	target, ok := request["target"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected a proxy target, but was empty.")
	}

	if expect := "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"; target["url"] != expect {
		t.Errorf("Expected proxy target url %s, but was %s", expect, target["url"])
	}

	if expect := "jack"; target["user"] != expect {
		t.Errorf("Expected proxy target username %s, but was %s", expect, target["user"])
	}

	if expect := "benimble"; target["password"] != expect {
		t.Errorf("Expected proxy target password %s, but was %s", expect, target["password"])
	}
}

func setupServer(status int, resp string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		//body, err := ioutil.ReadAll(r.Body)
		//if err == nil {
		//	fmt.Println(string(body))
		//}

		fmt.Fprintln(w, resp)
	}))
}

func setupPlugin(t *testing.T, conf string) telegraf.Input {
	table, err := toml.Parse([]byte(conf))
	if err != nil {
		t.Fatalf("Unable to parse config! %v", err)
	}

	for name := range table.Fields {
		object := table.Fields[name]
		switch name {
		case "jolokia2_exec_agent":
			plugin := JolokiaAgent{
				Metrics:               []MetricConfig{},
				DefaultFieldSeparator: ".",
			}

			if err := toml.UnmarshalTable(object.(*ast.Table), &plugin); err != nil {
				t.Fatalf("Unable to parse jolokia_agent plugin config! %v", err)
			}

			return &plugin

		case "jolokia2_exec_proxy":
			plugin := JolokiaProxy{
				Metrics:               []MetricConfig{},
				DefaultFieldSeparator: ".",
			}

			if err := toml.UnmarshalTable(object.(*ast.Table), &plugin); err != nil {
				t.Fatalf("Unable to parse jolokia_proxy plugin config! %v", err)
			}

			return &plugin
		}
	}

	return nil
}

