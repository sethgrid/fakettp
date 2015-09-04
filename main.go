package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

// StringSlice adheres to the flag Var interface, and allows for the -header flag to be reused
type StringSlice []string

type Config struct {
	ProxyHost string  `json:"proxy_host"`
	ProxyPort int     `json:"proxy_port"`
	Port      int     `json:"port"`
	Fakes     []*Fake `json:"fakes"`
}

type Fake struct {
	HyjackPath      string      `json:"hyjack"`
	Methods         StringSlice `json:"methods"`
	ResponseBody    string      `json:"body"`
	ResponseCode    int         `json:"code"`
	ResponseHeaders StringSlice `json:"headers"`
	ResponseTimeRaw string      `json:"time"`
	ResponseTime    time.Duration
}

var GlobalConfig *Config

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var ConfigPath string

	var Port int
	var ResponseCode int
	var ResponseTime time.Duration
	var ResponseBody string
	var ResponseHeaders StringSlice

	var Methods StringSlice
	var HyjackPath string
	var ProxyHost string
	var ProxyPort int

	flag.StringVar(&ConfigPath, "config", "", "json formatted conf file (see README at github.com/sethgrid/fakettp). If this flag is used, no other flags will be recognized.")

	flag.IntVar(&Port, "port", 5000, "set the port on which to listen")
	flag.IntVar(&ResponseCode, "code", 200, "set the http status code with which to respond")
	flag.DurationVar(&ResponseTime, "time", time.Millisecond*10, "set the response time, ex: 250ms or 1m5s")
	flag.StringVar(&ResponseBody, "body", "", "set the response body")
	flag.Var(&ResponseHeaders, "header", "headers, ex: 'Content-Type: application/json'. Multiple -header parameters allowed.")

	flag.Var(&Methods, "method", "used with the -hyjack route to limit hyjacking to the given http verb. Multiple -method parameters allowed.")
	flag.StringVar(&HyjackPath, "hyjack", "", "set the route you wish to hijack if using the reverse proxy host and port")
	flag.StringVar(&ProxyHost, "proxy_host", "http://0.0.0.0", "the host we will reverse proxy to (include protocol)")
	flag.IntVar(&ProxyPort, "proxy_port", 0, "the proxy port")
	flag.Parse()

	if ConfigPath != "" {
		// if we have a config file, use it for all config values
		data, err := ioutil.ReadFile(ConfigPath)
		if err != nil {
			log.Fatalf("%s", string(data))
		}
		config := &Config{}
		err = json.Unmarshal(data, &config)
		if err != nil {
			log.Fatalf("%v", err)
		}

		GlobalConfig = config

		// set all the response times from config file string to time.Duration
		for _, fake := range GlobalConfig.Fakes {
			if fake.ResponseTimeRaw == "" {
				continue
			}
			d, err := time.ParseDuration(fake.ResponseTimeRaw)
			if err != nil {
				log.Fatalf("%v", err)
			}
			fake.ResponseTime = d
		}

		log.Printf("starting on port :%d using config %s", config.Port, ConfigPath)

	} else {
		// if we did not get a config generated from a file, generate one from the
		// passed in flags
		config := &Config{}
		config.Port = Port
		config.ProxyHost = ProxyHost
		config.ProxyPort = ProxyPort

		fake := &Fake{}
		fake.ResponseHeaders = ResponseHeaders
		fake.HyjackPath = HyjackPath
		fake.ResponseBody = ResponseBody
		fake.ResponseCode = ResponseCode
		fake.ResponseTime = ResponseTime

		config.Fakes = []*Fake{fake}

		GlobalConfig = config

		log.Printf("starting on port :%d", config.Port)
	}

	http.HandleFunc("/", defaultHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", GlobalConfig.Port), nil)
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

	log.Printf("new request %s %s", r.Method, r.RequestURI)

	for _, fake := range GlobalConfig.Fakes {
		if willHyjack(r.Method, fake.Methods, r.URL.Path, fake.HyjackPath) {
			log.Printf("hyjacking route %s (waiting %s)", fake.HyjackPath, fake.ResponseTime.String())
			if fake.ResponseTime > 0 {
				<-time.Tick(fake.ResponseTime)
			}
			for _, header := range fake.ResponseHeaders {
				parts := strings.Split(header, ": ")

				if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
					key, value := parts[0], parts[1]
					log.Printf("setting header %s:%s", key, value)
					w.Header().Add(key, value)
				} else {
					log.Printf("skipping header %s (need a value on both sides of :)", header)
				}
			}
			w.WriteHeader(fake.ResponseCode)
			w.Write([]byte(fake.ResponseBody))
			log.Println("hyjack request complete")
			return
		}
	}
	// not hyjacking this time
	log.Println("proxying request")

	director := func(req *http.Request) {
		// handle both cases where we got `http://hostname` or `hostname`
		parts := strings.Split(GlobalConfig.ProxyHost, "://")
		var scheme string
		var host string
		if len(parts) == 1 {
			scheme = "http"
			host = fmt.Sprintf("%s:%d", parts[0], GlobalConfig.ProxyPort)
		} else if len(parts) >= 2 {
			scheme = parts[0]
			host = fmt.Sprintf("%s:%d", parts[1], GlobalConfig.ProxyPort)
		} else {
			log.Printf("issue splitting host on :// - %s", GlobalConfig.ProxyHost)
			return
		}

		req = r
		req.URL.Scheme = scheme
		req.URL.Host = host
	}

	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(w, r)
	log.Printf("proxy request complete")
}

// willHyjack returns true when we have a hyjack route that matches our request path,
// and takes into account the methods we want to hyjack
func willHyjack(requestMethod string, hyjackMethods StringSlice, requestPath string, hyjackRoute string) bool {
	isHyjack := false

	if hyjackRoute != "" {
		if len(hyjackMethods) == 0 {
			if hyjackRoute == requestPath {
				isHyjack = true
			}
		}
		for _, method := range hyjackMethods {
			if strings.ToUpper(method) == strings.ToUpper(requestMethod) && hyjackRoute == requestPath {
				isHyjack = true
			}
		}
	}

	return isHyjack
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
