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
	var Methods StringSlice
	var HyjackPath string
	var ProxyHost string
	var ProxyPort int
	var ProxyDelayTime time.Duration
	var IsRegex bool

	C := populateGlobalConfig(getSampleConfig(), Port, ResponseCode, ResponseTime, ResponseBody, ResponseHeaders, Methods, HyjackPath, ProxyHost, ProxyPort, ProxyDelayTime, IsRegex)

	// top level config values
	if got, want := C.Port, 5002; got != want {
		t.Error("got port %d, want %d", got, want)
	}
	if got, want := C.ProxyHost, "apid.docker"; got != want {
		t.Error("got proxy host %s, want %s", got, want)
	}
	if got, want := C.ProxyPort, 9092; got != want {
		t.Error("got proxy port %d, want %d", got, want)
	}
	if got, want := C.ProxyDelayTime, time.Millisecond*3; got != want {
		t.Error("got delay time of %s, want %s", got.String(), want.String())
	}

	// fakes
	if got, want := len(C.Fakes), 2; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d fakes, want %d", got, want)
	}

	// first fake
	if got, want := C.Fakes[0].HyjackPath, "/api/settings.json"; got != want {
		t.Error("got hyjack path %s, want %s", got, want)
	}
	if got, want := len(C.Fakes[0].Methods), 0; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d methods, want %d", got, want)
	}
	if got, want := C.Fakes[0].ResponseBody, ""; got != want {
		t.Error("got %s, want nothing", got)
	}
	if got, want := C.Fakes[0].ResponseCode, 500; got != want {
		t.Error("got response code %d, want %d", got, want)
	}
	if got, want := len(C.Fakes[0].ResponseHeaders), 0; got != want {
		t.Error("got %d response headers, want %d", got, want)
	}
	if got, want := C.Fakes[0].ResponseTime, time.Duration(0); got != want {
		t.Error("got resposne time %v, want %v", got, want)
	}

	// second fake
	if got, want := C.Fakes[1].HyjackPath, "/api/functions.json"; got != want {
		t.Error("got hyjack path %s, want %s", got, want)
	}
	if got, want := len(C.Fakes[1].Methods), 2; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d methods, want %d", got, want)
	}
	if got, want := C.Fakes[1].Methods[0], "GET"; got != want {
		t.Errorf("got method %s, want %d", got, want)
	}
	if got, want := C.Fakes[1].Methods[1], "POST"; got != want {
		t.Errorf("got method %s, want %d", got, want)
	}
	if got, want := C.Fakes[1].ResponseBody, `{"json":true}`; got != want {
		t.Error("got body %s, want nothing", got)
	}
	if got, want := C.Fakes[1].ResponseCode, 201; got != want {
		t.Error("got response code %d, want %d", got, want)
	}
	if got, want := len(C.Fakes[1].ResponseHeaders), 2; got != want {
		// must fatal to prevent nil reverence panics below
		t.Fatalf("got %d response headers, want %d", got, want)
	}
	if got, want := C.Fakes[1].ResponseHeaders[0], "Content-Type: application/json"; got != want {
		t.Error("got header %s, want %s", got, want)
	}
	if got, want := C.Fakes[1].ResponseHeaders[1], "Cache-Control: max-age=3600"; got != want {
		t.Error("got header %s, want %s", got, want)
	}
	if got, want := C.Fakes[1].ResponseTime, time.Millisecond*1015; got != want {
		t.Error("got resposne time %v, want %v", got, want)
	}
}

func TestConfigFromParameters(t *testing.T) {
	defaultHyjackTestSetup()

	// flag parameters
	var Port int = 5000
	var ResponseCode int = 201
	var ResponseTime time.Duration = time.Millisecond * 1015
	var ResponseBody string = `{"json":true}`
	var ResponseHeaders StringSlice = StringSlice{"Content-Type: application/json", "Cache-Control: max-age=3600"}
	var Methods StringSlice = StringSlice{"GET", "POST"}
	var HyjackPath string = "/api/functions.json"
	var ProxyHost string = "apid.docker"
	var ProxyPort int = 9092
	var ProxyDelayTime time.Duration
	var IsRegex bool

	emptyConfigData := []byte{}
	C := populateGlobalConfig(emptyConfigData, Port, ResponseCode, ResponseTime, ResponseBody, ResponseHeaders, Methods, HyjackPath, ProxyHost, ProxyPort, ProxyDelayTime, IsRegex)

	// top level config values
	if got, want := C.Port, 5000; got != want {
		t.Error("got port %d, want %d", got, want)
	}
	if got, want := C.ProxyHost, "apid.docker"; got != want {
		t.Error("got proxy host %s, want %s", got, want)
	}
	if got, want := C.ProxyPort, 9092; got != want {
		t.Error("got proxy port %d, want %d", got, want)
	}

	// fakes
	if got, want := len(C.Fakes), 1; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d fake, want %d", got, want)
	}

	if got, want := C.Fakes[0].HyjackPath, "/api/functions.json"; got != want {
		t.Error("got hyjack path %s, want %s", got, want)
	}
	if got, want := len(C.Fakes[0].Methods), 2; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d methods, want %d", got, want)
	}
	if got, want := C.Fakes[0].Methods[0], "GET"; got != want {
		t.Errorf("got method %s, want %d", got, want)
	}
	if got, want := C.Fakes[0].Methods[1], "POST"; got != want {
		t.Errorf("got method %s, want %d", got, want)
	}
	if got, want := C.Fakes[0].ResponseBody, `{"json":true}`; got != want {
		t.Error("got body %s, want nothing", got)
	}
	if got, want := C.Fakes[0].ResponseCode, 201; got != want {
		t.Error("got response code %d, want %d", got, want)
	}
	if got, want := len(C.Fakes[0].ResponseHeaders), 2; got != want {
		// must fatal to prevent nil reverence panics below
		t.Fatalf("got %d response headers, want %d", got, want)
	}
	if got, want := C.Fakes[0].ResponseHeaders[0], "Content-Type: application/json"; got != want {
		t.Error("got header %s, want %s", got, want)
	}
	if got, want := C.Fakes[0].ResponseHeaders[1], "Cache-Control: max-age=3600"; got != want {
		t.Error("got header %s, want %s", got, want)
	}
	if got, want := C.Fakes[0].ResponseTime, time.Millisecond*1015; got != want {
		t.Error("got resposne time %v, want %v", got, want)
	}
}

func TestConfigFromFileAndParameters(t *testing.T) {
	defaultHyjackTestSetup()

	// flag parameters
	var Port int = 5001
	var ResponseCode int = 201
	var ResponseTime time.Duration = time.Millisecond * 1015
	var ResponseBody string = `{"json":true}`
	var ResponseHeaders StringSlice = StringSlice{"Content-Type: application/json", "Cache-Control: max-age=3600"}
	var Methods StringSlice = StringSlice{"GET", "POST"}
	var HyjackPath string = "/api/functions.json"
	var ProxyHost string = "apid2.docker"
	var ProxyPort int = 9093
	var ProxyDelayTime time.Duration
	var IsRegex bool

	C := populateGlobalConfig(getSampleConfig(), Port, ResponseCode, ResponseTime, ResponseBody, ResponseHeaders, Methods, HyjackPath, ProxyHost, ProxyPort, ProxyDelayTime, IsRegex)

	// top level config values, config data overridden by parameters
	if got, want := C.Port, 5001; got != want {
		t.Error("got port %d, want %d", got, want)
	}
	if got, want := C.ProxyHost, "apid2.docker"; got != want {
		t.Error("got proxy host %s, want %s", got, want)
	}
	if got, want := C.ProxyPort, 9093; got != want {
		t.Error("got proxy port %d, want %d", got, want)
	}

	// fakes
	// get 2 from config and 1 from command line
	if got, want := len(C.Fakes), 3; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d fakes, want %d", got, want)
	}

	// first fake
	if got, want := C.Fakes[0].HyjackPath, "/api/settings.json"; got != want {
		t.Error("got hyjack path %s, want %s", got, want)
	}
	if got, want := len(C.Fakes[0].Methods), 0; got != want {
		// must fatal to prevent nil reference panics below
		t.Fatalf("got %d methods, want %d", got, want)
	}
	if got, want := C.Fakes[0].ResponseBody, ""; got != want {
		t.Error("got %s, want nothing", got)
	}
	if got, want := C.Fakes[0].ResponseCode, 500; got != want {
		t.Error("got response code %d, want %d", got, want)
	}
	if got, want := len(C.Fakes[0].ResponseHeaders), 0; got != want {
		t.Error("got %d response headers, want %d", got, want)
	}
	if got, want := C.Fakes[0].ResponseTime, time.Duration(0); got != want {
		t.Error("got resposne time %v, want %v", got, want)
	}

	// second and third fake
	for i := 1; i <= 2; i++ {
		if got, want := C.Fakes[i].HyjackPath, "/api/functions.json"; got != want {
			t.Error("got hyjack path %s, want %s", got, want)
		}
		if got, want := len(C.Fakes[i].Methods), 2; got != want {
			// must fatal to prevent nil reference panics below
			t.Fatalf("got %d methods, want %d", got, want)
		}
		if got, want := C.Fakes[i].Methods[0], "GET"; got != want {
			t.Errorf("got method %s, want %d", got, want)
		}
		if got, want := C.Fakes[i].Methods[1], "POST"; got != want {
			t.Errorf("got method %s, want %d", got, want)
		}
		if got, want := C.Fakes[i].ResponseBody, `{"json":true}`; got != want {
			t.Error("got body %s, want nothing", got)
		}
		if got, want := C.Fakes[i].ResponseCode, 201; got != want {
			t.Error("got response code %d, want %d", got, want)
		}
		if got, want := len(C.Fakes[i].ResponseHeaders), 2; got != want {
			// must fatal to prevent nil reverence panics below
			t.Fatalf("got %d response headers, want %d", got, want)
		}
		if got, want := C.Fakes[i].ResponseHeaders[0], "Content-Type: application/json"; got != want {
			t.Error("got header %s, want %s", got, want)
		}
		if got, want := C.Fakes[i].ResponseHeaders[1], "Cache-Control: max-age=3600"; got != want {
			t.Error("got header %s, want %s", got, want)
		}
		if got, want := C.Fakes[i].ResponseTime, time.Millisecond*1015; got != want {
			t.Error("got resposne time %v, want %v", got, want)
		}
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
        }
    ]
}`)
}
