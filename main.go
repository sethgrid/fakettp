package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var Port int
var ResponseCode int
var ResponseTime time.Duration
var ResponseBody string
var ResponseHeaders StringSlice

// StringSlice adheres to the flag Var interface, and allows for the -header flag to be reused
type StringSlice []string

func main() {
	flag.IntVar(&Port, "port", 5000, "set the port on which to listen")
	flag.IntVar(&ResponseCode, "code", 200, "set the http status code with which to respond")
	flag.DurationVar(&ResponseTime, "response_time", time.Millisecond*10, "set the response time, ex: 250ms or 1m5s")
	flag.StringVar(&ResponseBody, "body", "", "set the response body")
	flag.Var(&ResponseHeaders, "header", "headers, ex: 'Content-Type: application/json'")
	flag.Parse()

	log.Printf("starting on port :%d with response code %d and response time %s with headers %v", Port, ResponseCode, ResponseTime.String(), ResponseHeaders)

	http.HandleFunc("/", defaultHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("new request, waiting %s", ResponseTime.String())
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
