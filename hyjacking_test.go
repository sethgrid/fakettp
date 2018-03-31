package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
var serversStarted bool

// defaultHyjackTestSetup() is not concurrent test safe (use of hard coded ports and a global variable)
func defaultHyjackTestSetup() {
	if !*enableTestLogs {
		log.SetOutput(ioutil.Discard)
	}

	// flag parameters
	var Port = 4333
	var ResponseCode = http.StatusTeapot
	var ResponseTime time.Duration
	var ResponseBody = "hyjacked"
	var ResponseHeaders = StringSlice{"Cache-Control: max-age=3600"}
	var Methods = StringSlice{"GET"}
	var HyjackPath = "/bar"
	var ProxyHost = "127.0.0.1"
	var ProxyPort = 4332
	var ProxyDelayTime time.Duration
	var IsRegex bool
	var UseRequestURI bool

	GlobalConfig = populateGlobalConfig(getSampleConfig(), Port, ResponseCode, ResponseTime, ResponseBody, ResponseHeaders, Methods, HyjackPath, ProxyHost, ProxyPort, ProxyDelayTime, IsRegex, UseRequestURI)

	if !serversStarted {
		// start fakettp proxy and backing server
		go startFakettp(GlobalConfig.Port)
		go func() {
			err := http.ListenAndServe(fmt.Sprintf(":%d", ProxyPort), &testMux{})
			if err != nil {
				fmt.Printf("error creating backing server for hyjack and proxy test - %v", err)
				os.Exit(1)
			}
		}()
		serversStarted = true
	}
	// give the services time to start
	time.Tick(150 * time.Millisecond)
}

func TestBackingServerSetup(t *testing.T) {
	defaultHyjackTestSetup()

	t.Log(">> verify requests to backing server work")
	{
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/foo", GlobalConfig.ProxyHost, GlobalConfig.ProxyPort))
		if err != nil {
			t.Fatalf("error getting url from backing service - %v", err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		// while the response says "proxied", it is actually just a direct call
		if got, want := string(body), "proxied"; got != want {
			t.Errorf("\ngot body:\n%s\nwant body:\n%s\n", got, want)
		}
	}
}

func TestProxySetup(t *testing.T) {
	defaultHyjackTestSetup()
	t.Log(">> verify requests can be proxied")
	{
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/foo", GlobalConfig.ProxyPort))
		if err != nil {
			t.Fatalf("error getting url from proxy service - %v", err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if got, want := string(body), "proxied"; got != want {
			t.Errorf("\ngot body:\n%s\nwant body:\n%s\n", got, want)
		}
	}
}
func TestHyjacking(t *testing.T) {
	defaultHyjackTestSetup()
	t.Log(">> verify requests can be hyjacked")
	{
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/bar", GlobalConfig.Port))
		if err != nil {
			t.Fatalf("error getting url from proxy service - %v", err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if got, want := string(body), `hyjacked`; got != want {
			t.Errorf("\ngot body:\n%s\nwant body:\n%s\n", got, want)
		}
		if got, want := resp.StatusCode, http.StatusTeapot; got != want {
			t.Errorf("got status code %d, want %d", got, want)
		}
		if got, want := resp.Header.Get("Cache-Control"), "max-age=3600"; got != want {
			t.Errorf("got value for header Cache-Control `%s`, want `%s`", got, want)
		}
	}
}
func TestProxyBasedOnMethod(t *testing.T) {
	defaultHyjackTestSetup()
	t.Log(">> verify that only methods specified are hyjacked (post is not specified, should be proxied)")
	{
		resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/bar", GlobalConfig.Port), "application/json", strings.NewReader("body!"))
		if err != nil {
			t.Fatalf("error getting url from proxy service - %v", err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if got, want := string(body), `proxied`; got != want {
			t.Errorf("\ngot body:\n%s\nwant body:\n%s\n", got, want)
		}
	}
}
func TestPatternMatching(t *testing.T) {
	defaultHyjackTestSetup()
	t.Log(">> verify requests can be hyjacked using pattern matching routes")
	{
		GlobalConfig.Fakes[len(GlobalConfig.Fakes)-1].HyjackPath = `\/api\/users\/[0-9]+\/credits.json`
		GlobalConfig.Fakes[len(GlobalConfig.Fakes)-1].IsRegex = true

		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/users/1234/credits.json", GlobalConfig.Port))
		if err != nil {
			t.Fatalf("error getting url from proxy service - %v", err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if got, want := string(body), `hyjacked`; got != want {
			t.Errorf("\ngot body:\n%s\nwant body:\n%s\n", got, want)
		}
	}
}

func TestRequestURI(t *testing.T) {
	defaultHyjackTestSetup()
	t.Log(">> verify requests can be hyjacked using query param")
	{
		GlobalConfig.Fakes[len(GlobalConfig.Fakes)-1].HyjackPath = `\/api\/users\/[0-9]+\/credits\.json\?foo`
		GlobalConfig.Fakes[len(GlobalConfig.Fakes)-1].IsRegex = true
		GlobalConfig.Fakes[len(GlobalConfig.Fakes)-1].UseRequestURI = true

		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/users/1234/credits.json?foo", GlobalConfig.Port))
		if err != nil {
			t.Fatalf("error getting url from proxy service - %v", err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if got, want := string(body), `hyjacked`; got != want {
			t.Errorf("\ngot body:\n%s\nwant body:\n%s\n", got, want)
		}
	}
}
