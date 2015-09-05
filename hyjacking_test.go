package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

// testMux is for a backing service to show that proxying works
type testMux struct{}

func (m *testMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("proxied"))
}

var enableTestLogs = flag.Bool("show_logs", false, "`go test -show_logs` will enable application logging")

// TestHyjackAndProxy, not concurrent test safe (use of hard coded ports and a global variable)
func TestHyjackAndProxy(t *testing.T) {
	if !*enableTestLogs {
		log.SetOutput(ioutil.Discard)
	}

	// flag parameters
	var Port = 4333
	var ResponseCode int = http.StatusTeapot
	var ResponseTime time.Duration
	var ResponseBody string = "hyjacked"
	var ResponseHeaders StringSlice = StringSlice{"Cache-Control: max-age=3600"}
	var Methods StringSlice = StringSlice{"GET"}
	var HyjackPath string = "/bar"
	var ProxyHost string = "127.0.0.1"
	var ProxyPort int = 4332

	GlobalConfig = populateGlobalConfig(getSampleConfig(), Port, ResponseCode, ResponseTime, ResponseBody, ResponseHeaders, Methods, HyjackPath, ProxyHost, ProxyPort)

	// start fakettp proxy and backing server
	go startFakettp(GlobalConfig.Port)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", ProxyPort), &testMux{})
		if err != nil {
			t.Fatalf("error creating backing server for hyjack and proxy test - %v", err)
		}
	}()
	// give the services time to start
	time.Tick(150 * time.Millisecond)

	t.Log(">> verify requests to backing server work")
	{
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/foo", ProxyHost, ProxyPort))
		if err != nil {
			t.Fatalf("error getting url from backing service - %v", err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		// while the response says "proxied", it is actually just a direct call
		if got, want := string(body), "proxied"; got != want {
			t.Errorf("got body:\n%s\nwant body:\n%s\n`", got, want)
		}
	}

	t.Log(">> verify requests can be proxied")
	{
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/foo", ProxyPort))
		if err != nil {
			t.Fatalf("error getting url from proxy service - %v", err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if got, want := string(body), "proxied"; got != want {
			t.Errorf("got body:\n%s\nwant body:\n%s\n`", got, want)
		}
	}

	t.Log(">> verify requests can be hyjacked")
	{
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/bar", GlobalConfig.Port))
		if err != nil {
			t.Fatalf("error getting url from proxy service - %v", err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if got, want := string(body), `hyjacked`; got != want {
			t.Errorf("got body:\n%s\nwant body:\n%s\n`", got, want)
		}
		if got, want := resp.StatusCode, http.StatusTeapot; got != want {
			t.Errorf("got status code %d, want %d", got, want)
		}
		if got, want := resp.Header.Get("Cache-Control"), "max-age=3600"; got != want {
			t.Errorf("got value for header Cache-Control `%s`, want `%s`", got, want)
		}
	}

	t.Log(">> verify that only methods specified are hyjacked")
	{
		resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/bar", GlobalConfig.Port), "application/json", strings.NewReader(""))
		if err != nil {
			t.Fatalf("error getting url from proxy service - %v", err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if got, want := string(body), `proxied`; got != want {
			t.Errorf("got body:\n%s\nwant body:\n%s\n`", got, want)
		}
	}
}
