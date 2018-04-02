package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// StringSlice adheres to the flag Var interface, and allows for the -header flag to be reused
type StringSlice []string

type Config struct {
	ProxyHost      string  `json:"proxy_host"`
	ProxyPort      int     `json:"proxy_port"`
	Port           int     `json:"port"`
	Fakes          []*Fake `json:"fakes"`
	ProxyDelayRaw  string  `json:"proxy_delay"`
	ProxyDelayTime time.Duration
}

type Fake struct {
	HyjackPath        string      `json:"hyjack"`
	Methods           StringSlice `json:"methods"`
	RequestBodySubStr string      `json:"request_body"`
	ResponseBody      string      `json:"body"`
	ResponseCode      int         `json:"code"`
	ResponseHeaders   StringSlice `json:"headers"`
	ResponseTimeRaw   string      `json:"time"`
	IsRegex           bool        `json:"pattern_match"`
	UseRequestURI     bool        `json:"request_uri"`
	ResponseTime      time.Duration
}

func (f *Fake) String() string {
	var methods string
	if len(f.Methods) == 0 {
		methods = "[ALL METHODS]"
	} else {
		methods = fmt.Sprintf("%v", f.Methods)
	}

	var path string
	if len(f.HyjackPath) == 0 {
		path = "all paths"
	} else {
		path = f.HyjackPath
	}

	return fmt.Sprintf("fake: %s %s -> code %d, headers %v, time %s, body `%s`", methods, path, f.ResponseCode, f.ResponseHeaders, f.ResponseTime.String(), f.ResponseBody)
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
	var RequestBodySubStr string
	var IsRegex bool
	var UseRequestURI bool

	var HyjackPath string
	var ProxyHost string
	var ProxyPort int
	var ProxyDelayTime time.Duration

	flag.StringVar(&ConfigPath, "config", "", "json formatted conf file (see README at github.com/sethgrid/fakettp). If this flag is used, no other flags will be recognized.")

	flag.IntVar(&Port, "port", 0, "set the port on which to listen")
	flag.IntVar(&ResponseCode, "code", 0, "set the http status code with which to respond")
	flag.DurationVar(&ResponseTime, "time", time.Millisecond*0, "set the response time, ex: 250ms or 1m5s")
	flag.StringVar(&ResponseBody, "body", "", "set the response body")
	flag.StringVar(&RequestBodySubStr, "request_body_substr", "", "match against POST body with given substr")
	flag.Var(&ResponseHeaders, "header", "headers, ex: 'Content-Type: application/json'. Multiple -header parameters allowed.")
	flag.BoolVar(&IsRegex, "pattern_match", false, "set to true to match route patterns with Go regular expressions")
	flag.BoolVar(&UseRequestURI, "request_uri", false, "set to true to match on raw query (including query params)")
	flag.Var(&Methods, "method", "used with the -hyjack route to limit hyjacking to the given http verb. Multiple -method parameters allowed.")

	flag.StringVar(&HyjackPath, "hyjack", "", "set the route you wish to hijack if using the reverse proxy host and port")
	flag.StringVar(&ProxyHost, "proxy_host", "", "the host we will reverse proxy to (include protocol)")
	flag.IntVar(&ProxyPort, "proxy_port", 0, "the proxy port")
	flag.DurationVar(&ProxyDelayTime, "proxy_delay", time.Millisecond*0, "set the response time for proxied endpoints, ex: 250ms or 1m5s")
	flag.Parse()

	ConfigData := []byte{}
	var err error

	if ConfigPath != "" {
		ConfigData, err = ioutil.ReadFile(ConfigPath)
		if err != nil {
			log.Fatalf("%s", string(ConfigData))
		}
	}

	GlobalConfig = populateGlobalConfig(ConfigData, Port, ResponseCode, ResponseTime, ResponseBody, ResponseHeaders, Methods, RequestBodySubStr, HyjackPath, ProxyHost, ProxyPort, ProxyDelayTime, IsRegex, UseRequestURI)
	log.Printf("starting on port :%d", GlobalConfig.Port)

	startFakettp(GlobalConfig.Port)
}

func startFakettp(port int) {
	http.HandleFunc("/", defaultHandler)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func populateGlobalConfig(ConfigData []byte, Port int, ResponseCode int, ResponseTime time.Duration, ResponseBody string, ResponseHeaders StringSlice, Methods StringSlice, RequestBodySubStr string, HyjackPath string, ProxyHost string, ProxyPort int, ProxyDelayTime time.Duration, IsRegex, UseRequestURI bool) *Config {
	config := &Config{}

	if len(ConfigData) != 0 {
		err := json.Unmarshal(ConfigData, &config)
		if err != nil {
			log.Fatalf("parsing json error - %v", err)
		}

		if config.ProxyDelayRaw != "" {
			d, err := time.ParseDuration(config.ProxyDelayRaw)
			if err != nil {
				log.Fatalf("converting string delay to time duration - %v", err)
			}
			config.ProxyDelayTime = d
		}

		// set all the response times from config file string to time.Duration
		for _, fake := range config.Fakes {
			log.Printf("creating hyjack %s", fake)
			if fake.ResponseTimeRaw == "" {
				continue
			}
			d, err := time.ParseDuration(fake.ResponseTimeRaw)
			if err != nil {
				log.Fatalf("converting string delay to time duration - %v", err)
			}
			fake.ResponseTime = d
		}
	}

	// if we had command line values, use those too (override port and proxy settings)
	if Port != 0 {
		config.Port = Port
	} else if Port == 0 && config.Port == 0 {
		config.Port = 5000
	}
	if ProxyHost != "" {
		config.ProxyHost = ProxyHost
	}
	if ProxyPort != 0 {
		config.ProxyPort = ProxyPort
	}
	if ProxyDelayTime != 0 {
		config.ProxyDelayTime = ProxyDelayTime
	}

	// other rules on config
	if config.ProxyPort == 0 && config.Port != 0 {
		config.ProxyPort = config.Port
	}

	if len(config.Fakes) > 0 && HyjackPath != "" {
		log.Println("appending fake based on parameters")
		// if we are hyjacking a path beyond the config
		fake := &Fake{}
		fake.ResponseHeaders = ResponseHeaders
		fake.HyjackPath = HyjackPath
		fake.Methods = Methods
		fake.RequestBodySubStr = RequestBodySubStr
		fake.ResponseBody = ResponseBody
		fake.ResponseCode = ResponseCode
		fake.ResponseTime = ResponseTime
		fake.IsRegex = IsRegex
		fake.UseRequestURI = UseRequestURI
		config.Fakes = append(config.Fakes, fake)

	} else if len(ResponseHeaders) != 0 || HyjackPath != "" || ResponseCode != 0 || ResponseTime != 0 || len(Methods) != 0 {
		// no config fakes; if we have any parameters, let's use them
		log.Println("creating fake based on parameters")
		fake := &Fake{}
		fake.ResponseHeaders = ResponseHeaders
		fake.HyjackPath = HyjackPath
		fake.Methods = Methods
		fake.RequestBodySubStr = RequestBodySubStr
		fake.ResponseBody = ResponseBody
		fake.ResponseCode = ResponseCode
		fake.ResponseTime = ResponseTime
		fake.IsRegex = IsRegex
		fake.UseRequestURI = UseRequestURI
		log.Printf("creating hyjack %s", fake)
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

	// there are two ways that a request gets hyjacked:
	// 1 - X-Return-* header
	// 2 - Config
	// An X-Return-* header always overrides config.
	requestHyjacked := false
	var delay time.Duration
	var code int
	var headers http.Header
	var data []byte
	var err error

	if hdr := r.Header.Get("X-Return-Delay"); hdr != "" {
		delay, err = time.ParseDuration(hdr)
		if err != nil {
			log.Println("cannot set delay", err)
		}
	}
	// respect config delay if it was not set by header
	if delay == 0 && GlobalConfig.ProxyDelayTime > 0 {
		delay = GlobalConfig.ProxyDelayTime
	}

	if hdr := r.Header.Get("X-Return-Headers"); hdr != "" {
		requestHyjacked = true
		err = json.Unmarshal([]byte(hdr), &headers)
		if err != nil {
			requestHyjacked = false
			log.Println("unable to read X-Return-Headers", err)
		}
		for k, vs := range headers {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
	}
	if hdr := r.Header.Get("X-Return-Code"); hdr != "" {
		requestHyjacked = true
		code, err = strconv.Atoi(hdr)
		if err != nil {
			requestHyjacked = false
			log.Println("unable to read X-Return-Code", err)
		}
	}
	if hdr := r.Header.Get("X-Return-Data"); hdr != "" {
		requestHyjacked = true
		data = []byte(hdr)
	}

	if requestHyjacked {
		log.Printf("hyjacking request %s (waiting %s)", r.RequestURI, delay.String())
		w.WriteHeader(code)
		for name, values := range headers {
			log.Printf("setting header %s:%s", name, strings.Join(values, ","))
			w.Header().Set(name, strings.Join(values, ","))
		}
		time.Sleep(delay)
		w.Write(data)
		log.Println("hyjack X-Return-* request complete")
		return
	}

	// If this request was not X-Return-* based, check config.
	// Range over the configured fakes and determine if we
	// should hyjack the route
	for _, fake := range GlobalConfig.Fakes {
		pathToMatch := r.URL.Path
		if fake.UseRequestURI {
			pathToMatch = r.RequestURI
		}

		// extract the reqeust body, and put it back into the request.
		var originalRequestBody []byte
		var err error
		if r.Body != nil {
			originalRequestBody, err = ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("unable to read original request body - %v", err)
			}
			r.Body.Close()
		}

		// rehydrate the body (it is drained each read)
		if len(originalRequestBody) > 0 {
			r.Body = ioutil.NopCloser(bytes.NewBuffer(originalRequestBody))
		}

		if willHyjack(r.Method, fake.Methods, pathToMatch, fake.HyjackPath, string(originalRequestBody), fake.RequestBodySubStr, fake.IsRegex) {
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

	if delay > 0 {
		log.Printf("delaying proxy request %s", delay.String())
		time.Sleep(delay)
	}

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

		log.Println("setting scheme as ", scheme)
		req.URL.Scheme = scheme
		req.URL.Host = host
	}

	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(w, r)
	log.Printf("proxy request complete")
}

// willHyjack returns true when we have a hyjack route that matches our request path,
// and takes into account the methods we want to hyjack
func willHyjack(requestMethod string, hyjackMethods StringSlice, requestPath string, hyjackRoute string, requestBody string, requestBodySubStr string, isRegex bool) bool {
	methodMatches := false
	routeMatches := false
	requestBodyMatches := false
	requireBodyMatch := false

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

	if requestBodySubStr != "" {
		requireBodyMatch = true
	}

	if requestBodySubStr != "" && strings.Contains(requestBody, requestBodySubStr) {
		requestBodyMatches = true
	}

	if len(hyjackMethods) == 0 {
		methodMatches = true
	}
	for _, method := range hyjackMethods {
		if strings.ToUpper(method) == strings.ToUpper(requestMethod) {
			methodMatches = true
		}
	}

	if requireBodyMatch {
		return routeMatches && requestBodyMatches
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
