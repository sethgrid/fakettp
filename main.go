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
	"regexp"
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
	IsRegex         bool        `json:"pattern_match"`
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
	var IsRegex bool

	var HyjackPath string
	var ProxyHost string
	var ProxyPort int

	flag.StringVar(&ConfigPath, "config", "", "json formatted conf file (see README at github.com/sethgrid/fakettp). If this flag is used, no other flags will be recognized.")

	flag.IntVar(&Port, "port", 5000, "set the port on which to listen")
	flag.IntVar(&ResponseCode, "code", 200, "set the http status code with which to respond")
	flag.DurationVar(&ResponseTime, "time", time.Millisecond*10, "set the response time, ex: 250ms or 1m5s")
	flag.StringVar(&ResponseBody, "body", "", "set the response body")
	flag.Var(&ResponseHeaders, "header", "headers, ex: 'Content-Type: application/json'. Multiple -header parameters allowed.")
	flag.BoolVar(&IsRegex, "pattern_match", false, "set to true to match route patterns with Go regular expressions")
	flag.Var(&Methods, "method", "used with the -hyjack route to limit hyjacking to the given http verb. Multiple -method parameters allowed.")

	flag.StringVar(&HyjackPath, "hyjack", "", "set the route you wish to hijack if using the reverse proxy host and port")
	flag.StringVar(&ProxyHost, "proxy_host", "http://0.0.0.0", "the host we will reverse proxy to (include protocol)")
	flag.IntVar(&ProxyPort, "proxy_port", 0, "the proxy port")
	flag.Parse()

	ConfigData := []byte{}
	var err error

	if ConfigPath != "" {
		ConfigData, err = ioutil.ReadFile(ConfigPath)
		if err != nil {
			log.Fatalf("%s", string(ConfigData))
		}
	}

	GlobalConfig = populateGlobalConfig(ConfigData, Port, ResponseCode, ResponseTime, ResponseBody, ResponseHeaders, Methods, HyjackPath, ProxyHost, ProxyPort, IsRegex)
	log.Printf("starting on port :%d", GlobalConfig.Port)

	startFakettp(GlobalConfig.Port)
}

func startFakettp(port int) {
	http.HandleFunc("/", defaultHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func populateGlobalConfig(ConfigData []byte, Port int, ResponseCode int, ResponseTime time.Duration, ResponseBody string, ResponseHeaders StringSlice, Methods StringSlice, HyjackPath string, ProxyHost string, ProxyPort int, IsRegex bool) *Config {
	config := &Config{}
	// config.Fakes = []*Fake{}

	if len(ConfigData) != 0 {
		err := json.Unmarshal(ConfigData, &config)
		if err != nil {
			log.Fatalf("%v", err)
		}

		// set all the response times from config file string to time.Duration
		for _, fake := range config.Fakes {
			if fake.ResponseTimeRaw == "" {
				continue
			}
			d, err := time.ParseDuration(fake.ResponseTimeRaw)
			if err != nil {
				log.Fatalf("%v", err)
			}
			fake.ResponseTime = d
		}
	}

	// if we had command line values, use those too (override port and proxy settings)
	if Port != 0 {
		config.Port = Port
	}
	if ProxyHost != "" {
		config.ProxyHost = ProxyHost
	}
	if ProxyPort != 0 {
		config.ProxyPort = ProxyPort
	}

	// if there is data for a fake, grab it
	if len(ResponseHeaders) != 0 || HyjackPath != "" || ResponseCode != 0 || ResponseCode != 0 || ResponseTime != 0 || len(Methods) != 0 {
		fake := &Fake{}
		fake.ResponseHeaders = ResponseHeaders
		fake.HyjackPath = HyjackPath
		fake.Methods = Methods
		fake.ResponseBody = ResponseBody
		fake.ResponseCode = ResponseCode
		fake.ResponseTime = ResponseTime
		fake.IsRegex = IsRegex
		config.Fakes = append(config.Fakes, fake)
	}

	return config
}

// defaultHanlder will either proxy the request or substitute in the hyjack data
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	// capture a request id with padding and leading zeros incase multiple requests
	// come in at the same time
	reqID := fmt.Sprintf("[%07x] ", rand.Int31n(1e8))
	log.SetPrefix(reqID)

	log.Printf("new request %s %s", r.Method, r.RequestURI)

	for _, fake := range GlobalConfig.Fakes {
		if willHyjack(r.Method, fake.Methods, r.URL.Path, fake.HyjackPath, fake.IsRegex) {
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
func willHyjack(requestMethod string, hyjackMethods StringSlice, requestPath string, hyjackRoute string, isRegex bool) bool {
	methodMatches := false
	routeMatches := false

	if hyjackRoute == "" {
		routeMatches = true
	} else {
		if isRegex {
			var regex = regexp.MustCompile(hyjackRoute)
			if regex.MatchString(requestPath) {
				routeMatches = true
			}
		} else if hyjackRoute == requestPath {
			routeMatches = true
		}
	}

	if len(hyjackMethods) == 0 {
		methodMatches = true
	}
	for _, method := range hyjackMethods {
		if strings.ToUpper(method) == strings.ToUpper(requestMethod) {
			methodMatches = true
		}
	}

	return methodMatches && routeMatches
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
