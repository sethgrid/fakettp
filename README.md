Fakettp
--------

Fakettp is an http debugging proxy allowing you to easily black box test how an application reacts when services it relies upon act differently than expected. Fakettp allows you to control the following as you test at the black box level:
  - response code
  - response time
  - response headers
  - response body
  - control over which endpoints proxy though and which are hijacked

You can set global responses, or you can pass in individual endpoints where those responses apply, and proxy all other endpoint requests to the original service. This gives you the ability to test your system against the other service and simulate an error on a single endpoint while the rest of the service works normally.

How To Install
--------
```bash
go get github.com/sethgrid/fakettp
```

Use Case
--------

What inspired this repo!: We had Service A calling Service B's endpoints internally. We needed some of the endpoints in Service B to do their job, while for one of them we needed to verify how Service A behaved when Service B returned different content and status codes. After using this tool, our QA was able to verify all the edge-cases with which they were interested!

Sample Usage
------------

There are three main ways to use `fakeTTP`. Command line arguments, config file, or `X-Return-*` headers.

From the source or compiled binary, just set the response you want. Note that you can pass in multiple headers to be returned by using `-header` repeatedly. You can also limit which methods the hyjacking will affect with multiple `-method` parameters. You can also choose to match against a pattern string if you additionally pass in the `-pattern_match` flag.


Hyjac a single endpoint and proxy all other calls:
```
go run main.go -proxy_host http://example.com -proxy_port 9092 -hyjack /api/user -port 5555 -code 418 -body "I'm a teapot"
```

We can also use a config file instead (more on this below!) and override proxy settings with flags (or add an additional hyjacked route):
```
go run main.go -config my.conf -proxy_delay 5s
```

We can match against regular expressions / patterns:
```
$ go run main.go -port 5555 -code 201 -header POST -pattern_match -hyjack '\/api\/users\/[0-9]+\/credits.json'
```

We can use raw query including query params:
```
$ go run main.go -port 5555 -code 201 -header POST -pattern_match -request_uri -hyjack '\/api\/users\/[0-9]+\/credits.json\?foo=bar'
```

Return the same data for all calls to this service:
```
$ go run main.go -port 5555 -code 201 -header 'Content-Type: application/json' -header 'Cache-Control: max-age=3600' -body '{"json":true} -method POST -method GET'
```

The above setup would result in the following response from any path on `localhost:5555`:
```
$ curl -v localhost:5555/foo/bar/raz -d '{"foo":"bar"}'
*   Trying 127.0.0.1...
* Connected to localhost (127.0.0.1) port 5555 (#0)
> POST /foo/bar/raz HTTP/1.1
> Host: localhost:5555
> User-Agent: curl/7.43.0
> Accept: */*
> Content-Length: 13
> Content-Type: application/x-www-form-urlencoded
>
* upload completely sent off: 13 out of 13 bytes
< HTTP/1.1 201 Created
< Cache-Control: max-age=3600
< Content-Type: application/json
< Date: Wed, 02 Sep 2015 16:17:48 GMT
< Content-Length: 13
<
* Connection #0 to host localhost left intact
{"json":true}
```

Config File
-----------

When passing command line flags, you are limited to either hyjacking all requests or only requests to a single endpoint. With a config file, you can specify multiple routes to behave differently. Note: config values are overridden by command line flags in the case of `proxy_host`, `proxy_port`, and `port`. For all other values, they add an additional fake for hyjacking.

In the fakes list, you can set the "hyjack" url that will be matched against. For return values, you can specify code, body, headers, and time to delay the response.

There are some additional configs that deal with the matching. You can specify that the hyjack url is intended for a pattern_match (using standard regex). Normally, the hyjack url will just match the URL.path. If you request_uri to be true, it will match against the request's RequestURI. Lastly, for matching against different POST requests where the urls will be the same, you can specify the request_body param which will match if the given substring is in the request body payload.

Sample Config:
```json
{
    "proxy_host": "http://0.0.0.0",
    "proxy_port": 9092,
    "proxy_delay": "2s",
    "port": 5000,
    "fakes": [
        {
            "hyjack": "/api/settings.json",
            "code": 500
        },
        {
            "hyjack": "/api/users.json",
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
            "time": "1m3s15ms"
        },
        {
            "hyjack": "\/api\/users\/[0-9]+\/credits.json",
            "code": 404,
            "pattern_match": true
        },
        {
            "hyjack": "\/api\/users\/[0-9]+\/credits.json\\?foo=bar",
            "code": 404,
            "pattern_match": true,
            "request_uri": true
        },
        {
            "hyjack": "/api/post",
            "methods": [
                "POST"
            ],
            "code":200,
            "body": "hyjacked",
            "request_body": "catch me"
	}
    ]
}
```

X-Return-* Headers
-----------
You can hit the proxy directly and bypass configurations by using the following `X-Return-*` headers. 
 - X-Return-Delay: a time parsable duration, like 250ms or 1m30s.
 - X-Return-Code: a valid http status code
 - X-Return-Data: the string data you'd like to return
 - X-Return-Headers: a json blob of `map[string][]string`, such as `{"X-Custom-Header":["custom value"]}`.

This allows you to use the `fakeTTP` binary in a more programatic fashion.

Docker Use Cases
-----------
You can also use this in docker-compose like so,

```yaml
# fakettp: hijack certain HTTP routes / methods
# for instance, below command will hijack all DELETE calls and return 400, passing everything else to ceph-demo.dfs.docker
# command: fakettp -proxy_host ceph-demo.dfs.docker -proxy_port 80 -port 80 -method DELETE -code 400 -body ""
# you can also use volume mount to use a json config directly

fakettp:
  image: docker.sendgrid.net/sethgrid/fakettp
  command: fakettp -config /mount/fakettp.json
  links:
    - ceph-demo
  dns:
    - 172.17.0.1
  dns_search:
    - dfs.docker
  volumes:
    - .:/mount
```

Sample Logs
-----------

Logs show the requested URI, if the request was hyjacked or proxied, and shows a request id at the beginning of log lines dealing with a request.

```
2015/09/02 14:10:22 starting on port :5555
[1dfc34e] 2015/09/02 14:10:23 new request /api/user/get.json
[1dfc34e] 2015/09/02 14:10:23 proxying request
[4666fb4] 2015/09/02 14:10:36 new request /api/functions.json
[4666fb4] 2015/09/02 14:10:36 hyjacking route /api/functions.json (10ms)
[4666fb4] 2015/09/02 14:10:36 setting header Content-Type:application/json
[4666fb4] 2015/09/02 14:10:36 setting header Cache-Control:max-age=3600
```

Tests
-----

You can run tests like normal with `$ go test`, however, you can make the tests show application level logging with `$ go test -show_logs`. Tests make use of ports `4333` and `4332` and require these ports to be available.
