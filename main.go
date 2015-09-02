package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var Port int
var HyjackRoute string
var ResponseCode int
var ResponseTime time.Duration
var ResponseBody string
var ResponseHeaders StringSlice

var ProxyHost string
var ProxyPort int

// StringSlice adheres to the flag Var interface, and allows for the -header flag to be reused
type StringSlice []string

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.IntVar(&Port, "port", 5000, "set the port on which to listen")
	flag.IntVar(&ResponseCode, "code", 200, "set the http status code with which to respond")
	flag.DurationVar(&ResponseTime, "response_time", time.Millisecond*10, "set the response time, ex: 250ms or 1m5s")
	flag.StringVar(&ResponseBody, "body", "", "set the response body")
	flag.StringVar(&HyjackRoute, "hyjack", "", "set the route you wish to hijack if using the reverse proxy host and port")
	flag.StringVar(&ProxyHost, "proxy_host", "http://0.0.0.0", "the host we will reverse proxy to (include protocol)")
	flag.IntVar(&ProxyPort, "proxy_port", 0, "the proxy port")
	flag.Var(&ResponseHeaders, "header", "headers, ex: 'Content-Type: application/json'. Multiple -header parameters allowed.")
	flag.Parse()

	log.Printf("starting on port :%d", Port)

	http.HandleFunc("/", defaultHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

// defaultHanlder will either proxy the request or substitute in the hyjack data
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	// capture a request id with padding and leading zeros incase multiple requests
	// come in at the same time
	reqID := fmt.Sprintf("[%07x] ", rand.Int31n(1e8))
	log.SetPrefix(reqID)

	log.Printf("new request %s", r.RequestURI)

	if HyjackRoute == "" || HyjackRoute == r.URL.String() {
		log.Printf("hyjacking route %s (%s)", HyjackRoute, ResponseTime.String())
		<-time.Tick(ResponseTime)

		for _, header := range ResponseHeaders {
			parts := strings.Split(header, ": ")

			if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
				key, value := parts[0], parts[1]
				log.Printf("setting header %s:%s", key, value)
				w.Header().Add(key, value)
			} else {
				log.Printf("skipping header %s (need a value on both sides of :)", header)
			}
		}
		w.WriteHeader(ResponseCode)
		w.Write([]byte(ResponseBody))
		return
	}

	// not hyjacking this time
	log.Println("proxying request")

	req, err := http.NewRequest(r.Method, fmt.Sprintf("%s:%d%s", ProxyHost, ProxyPort, r.RequestURI), r.Body)
	if err != nil {
		log.Printf("error with proxy request - %v", err)
		return
	}

	for k, values := range r.Header {
		for _, value := range values {
			req.Header.Add(k, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("error with proxy request - %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	for k, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(k, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// String adheres to the flag Var interface
func (s *StringSlice) String() string {
	return fmt.Sprintf("%s", *s)
}

// Set adheres to the flag Var interface
func (s *StringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}
