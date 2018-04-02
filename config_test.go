package main

import (
	"testing"
	"time"
)

func TestConfigFromFile(t *testing.T) {
	defaultHyjackTestSetup()

	// flag parameters
	var Port int
	var ResponseCode int
	var ResponseTime time.Duration
	var ResponseBody string
	var ResponseHeaders StringSlice
	var RequestBodySubStr string
	var Methods StringSlice
	var HyjackPath string
	var ProxyHost string
	var ProxyPort int
	var ProxyDelayTime time.Duration
	var IsRegex bool
	var UseRequestURI bool

	C := populateGlobalConfig(getSampleConfig(), Port, ResponseCode, ResponseTime, ResponseBody, ResponseHeaders, Methods, RequestBodySubStr, HyjackPath, ProxyHost, ProxyPort, ProxyDelayTime, IsRegex, UseRequestURI)

	// top level config values
	if got, want := C.Port, 5002; got != want {
		t.Errorf("got port %d, want %d", got, want)
	}
	if got, want := C.ProxyHost, "apid.docker"; got != want {
		t.Errorf("got proxy host %s, want %s", got, want)
	}
	if got, want := C.ProxyPort, 9092; got != want {
		t.Errorf("got proxy port %d, want %d", got, want)
	}
	if got, want := C.ProxyDelayTime, time.Millisecond*3; got != want {
		t.Errorf("got delay time of %s, want %s", got.String(), want.String())
	}

	// fakes
	if got, want := len(C.Fakes), 3; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d fakes, want %d", got, want)
	}

	// first fake
	if got, want := C.Fakes[0].HyjackPath, "/api/settings.json"; got != want {
		t.Errorf("got hyjack path %s, want %s", got, want)
	}
	if got, want := len(C.Fakes[0].Methods), 0; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d methods, want %d", got, want)
	}
	if got, want := C.Fakes[0].ResponseBody, ""; got != want {
		t.Errorf("got %s, want nothing", got)
	}
	if got, want := C.Fakes[0].ResponseCode, 500; got != want {
		t.Errorf("got response code %d, want %d", got, want)
	}
	if got, want := len(C.Fakes[0].ResponseHeaders), 0; got != want {
		t.Errorf("got %d response headers, want %d", got, want)
	}
	if got, want := C.Fakes[0].ResponseTime, time.Duration(0); got != want {
		t.Errorf("got resposne time %v, want %v", got, want)
	}

	// second fake
	if got, want := C.Fakes[1].HyjackPath, "/api/functions.json"; got != want {
		t.Errorf("got hyjack path %s, want %s", got, want)
	}
	if got, want := len(C.Fakes[1].Methods), 2; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d methods, want %d", got, want)
	}
	if got, want := C.Fakes[1].Methods[0], "GET"; got != want {
		t.Errorf("got method %s, want %s", got, want)
	}
	if got, want := C.Fakes[1].Methods[1], "POST"; got != want {
		t.Errorf("got method %s, want %s", got, want)
	}
	if got, want := C.Fakes[1].ResponseBody, `{"json":true}`; got != want {
		t.Errorf("got body %s, want nothing", got)
	}
	if got, want := C.Fakes[1].ResponseCode, 201; got != want {
		t.Errorf("got response code %d, want %d", got, want)
	}
	if got, want := len(C.Fakes[1].ResponseHeaders), 2; got != want {
		// must fatal to prevent nil reverence panics below
		t.Fatalf("got %d response headers, want %d", got, want)
	}
	if got, want := C.Fakes[1].ResponseHeaders[0], "Content-Type: application/json"; got != want {
		t.Errorf("got header %s, want %s", got, want)
	}
	if got, want := C.Fakes[1].ResponseHeaders[1], "Cache-Control: max-age=3600"; got != want {
		t.Errorf("got header %s, want %s", got, want)
	}
	if got, want := C.Fakes[1].ResponseTime, time.Millisecond*1015; got != want {
		t.Errorf("got resposne time %v, want %v", got, want)
	}
}

func TestConfigFromParameters(t *testing.T) {
	defaultHyjackTestSetup()

	// flag parameters
	var Port = 5000
	var ResponseCode = 201
	var ResponseTime = time.Millisecond * 1015
	var ResponseBody = `{"json":true}`
	var ResponseHeaders = StringSlice{"Content-Type: application/json", "Cache-Control: max-age=3600"}
	var Methods = StringSlice{"GET", "POST"}
	var RequestBodySubStr string
	var HyjackPath = "/api/functions.json"
	var ProxyHost = "apid.docker"
	var ProxyPort = 9092
	var ProxyDelayTime time.Duration
	var IsRegex bool
	var UseRequestURI bool

	emptyConfigData := []byte{}
	C := populateGlobalConfig(emptyConfigData, Port, ResponseCode, ResponseTime, ResponseBody, ResponseHeaders, Methods, RequestBodySubStr, HyjackPath, ProxyHost, ProxyPort, ProxyDelayTime, IsRegex, UseRequestURI)

	// top level config values
	if got, want := C.Port, 5000; got != want {
		t.Errorf("got port %d, want %d", got, want)
	}
	if got, want := C.ProxyHost, "apid.docker"; got != want {
		t.Errorf("got proxy host %s, want %s", got, want)
	}
	if got, want := C.ProxyPort, 9092; got != want {
		t.Errorf("got proxy port %d, want %d", got, want)
	}

	// fakes
	if got, want := len(C.Fakes), 1; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d fake, want %d", got, want)
	}

	if got, want := C.Fakes[0].HyjackPath, "/api/functions.json"; got != want {
		t.Errorf("got hyjack path %s, want %s", got, want)
	}
	if got, want := len(C.Fakes[0].Methods), 2; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d methods, want %d", got, want)
	}
	if got, want := C.Fakes[0].Methods[0], "GET"; got != want {
		t.Errorf("got method %s, want %s", got, want)
	}
	if got, want := C.Fakes[0].Methods[1], "POST"; got != want {
		t.Errorf("got method %s, want %s", got, want)
	}
	if got, want := C.Fakes[0].ResponseBody, `{"json":true}`; got != want {
		t.Errorf("got body %s, want nothing", got)
	}
	if got, want := C.Fakes[0].ResponseCode, 201; got != want {
		t.Errorf("got response code %d, want %d", got, want)
	}
	if got, want := len(C.Fakes[0].ResponseHeaders), 2; got != want {
		// must fatal to prevent nil reverence panics below
		t.Fatalf("got %d response headers, want %d", got, want)
	}
	if got, want := C.Fakes[0].ResponseHeaders[0], "Content-Type: application/json"; got != want {
		t.Errorf("got header %s, want %s", got, want)
	}
	if got, want := C.Fakes[0].ResponseHeaders[1], "Cache-Control: max-age=3600"; got != want {
		t.Errorf("got header %s, want %s", got, want)
	}
	if got, want := C.Fakes[0].ResponseTime, time.Millisecond*1015; got != want {
		t.Errorf("got resposne time %v, want %v", got, want)
	}
}

func TestConfigFromFileAndParameters(t *testing.T) {
	defaultHyjackTestSetup()

	// flag parameters
	var Port = 5001
	var ResponseCode = 201
	var ResponseTime = time.Millisecond * 1015
	var ResponseBody = `{"json":true}`
	var ResponseHeaders = StringSlice{"Content-Type: application/json", "Cache-Control: max-age=3600"}
	var RequestBodySubStr string
	var Methods = StringSlice{"GET", "POST"}
	var HyjackPath = "/api/functions.json"
	var ProxyHost = "apid2.docker"
	var ProxyPort = 9093
	var ProxyDelayTime time.Duration
	var IsRegex bool
	var UseRequestURI bool

	C := populateGlobalConfig(getSampleConfig(), Port, ResponseCode, ResponseTime, ResponseBody, ResponseHeaders, Methods, RequestBodySubStr, HyjackPath, ProxyHost, ProxyPort, ProxyDelayTime, IsRegex, UseRequestURI)

	// top level config values, config data overridden by parameters
	if got, want := C.Port, 5001; got != want {
		t.Errorf("got port %d, want %d", got, want)
	}
	if got, want := C.ProxyHost, "apid2.docker"; got != want {
		t.Errorf("got proxy host %s, want %s", got, want)
	}
	if got, want := C.ProxyPort, 9093; got != want {
		t.Errorf("got proxy port %d, want %d", got, want)
	}

	// fakes
	// get 2 from config and 1 from command line
	if got, want := len(C.Fakes), 4; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d fakes, want %d", got, want)
	}

	// first fake
	if got, want := C.Fakes[0].HyjackPath, "/api/settings.json"; got != want {
		t.Errorf("got hyjack path %s, want %s", got, want)
	}
	if got, want := len(C.Fakes[0].Methods), 0; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d methods, want %d", got, want)
	}
	if got, want := C.Fakes[0].ResponseBody, ""; got != want {
		t.Errorf("got %s, want nothing", got)
	}
	if got, want := C.Fakes[0].ResponseCode, 500; got != want {
		t.Errorf("got response code %d, want %d", got, want)
	}
	if got, want := len(C.Fakes[0].ResponseHeaders), 0; got != want {
		t.Errorf("got %d response headers, want %d", got, want)
	}
	if got, want := C.Fakes[0].ResponseTime, time.Duration(0); got != want {
		t.Errorf("got resposne time %v, want %v", got, want)
	}

	if got, want := C.Fakes[1].HyjackPath, "/api/functions.json"; got != want {
		t.Errorf("got hyjack path %s, want %s", got, want)
	}
	if got, want := len(C.Fakes[1].Methods), 2; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d methods, want %d", got, want)
	}
	if got, want := C.Fakes[1].Methods[0], "GET"; got != want {
		t.Errorf("got method %s, want %s", got, want)
	}
	if got, want := C.Fakes[1].Methods[1], "POST"; got != want {
		t.Errorf("got method %s, want %s", got, want)
	}
	if got, want := C.Fakes[1].ResponseBody, `{"json":true}`; got != want {
		t.Errorf("got body %s, want nothing", got)
	}
	if got, want := C.Fakes[1].ResponseCode, 201; got != want {
		t.Errorf("got response code %d, want %d", got, want)
	}
	if got, want := len(C.Fakes[1].ResponseHeaders), 2; got != want {
		// must fatal to prevent nil reverence panics below
		t.Fatalf("got %d response headers, want %d", got, want)
	}
	if got, want := C.Fakes[1].ResponseHeaders[0], "Content-Type: application/json"; got != want {
		t.Errorf("got header %s, want %s", got, want)
	}
	if got, want := C.Fakes[1].ResponseHeaders[1], "Cache-Control: max-age=3600"; got != want {
		t.Errorf("got header %s, want %s", got, want)
	}
	if got, want := C.Fakes[1].ResponseTime, time.Millisecond*1015; got != want {
		t.Errorf("got resposne time %v, want %v", got, want)
	}

	if got, want := C.Fakes[2].HyjackPath, "/api/post"; got != want {
		t.Errorf("got hyjack path %s, want %s", got, want)
	}
	if got, want := len(C.Fakes[2].Methods), 1; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d methods, want %d", got, want)
	}
	if got, want := C.Fakes[2].Methods[0], "POST"; got != want {
		t.Errorf("got method %s, want %s", got, want)
	}
	if got, want := C.Fakes[2].ResponseBody, `hyjacked`; got != want {
		t.Errorf("got body %s, want nothing", got)
	}
	if got, want := C.Fakes[2].ResponseCode, 200; got != want {
		t.Errorf("got response code %d, want %d", got, want)
	}
	if got, want := len(C.Fakes[2].ResponseHeaders), 0; got != want {
		// must fatal to prevent nil reverence panics below
		t.Fatalf("got %d response headers, want %d", got, want)
	}
}

func getSampleConfig() []byte {
	return []byte(`{
    "proxy_host": "apid.docker",
    "proxy_port": 9092,
    "proxy_delay": "3ms",
    "port": 5002,
    "fakes": [
        {
            "hyjack": "/api/settings.json",
            "code": 500
        },
        {
            "hyjack": "/api/functions.json",
            "methods": [
                "GET",
                "POST"
            ],
            "body": "{\"json\":true}",
            "code": 201,
            "headers": [
                "Content-Type: application/json",
                "Cache-Control: max-age=3600"
            ],
            "time": "1s15ms"
        },{
			"hyjack": "/api/post",
			"methods": [
				"POST"
			],
			"code":200,
			"body": "hyjacked",
			"request_body": "catch me"
		}
    ]
}`)
}
