Fakettp
--------

Fakettp is an http debugging proxy allowing you to easily set up responses for http endpoints while testing. Fakettp allows you to control the following:
  - response code
  - response time
  - response headers
  - response body

You can have the above as global responses, or you can pass in an individual endpoint where those responses apply, and proxy all other endpoints to the original service. This gives you the ability to test your system against the other service and simulate an error on a single endpoint while the rest of the service works normally.

Use Case
--------

As a developer, you want to test how your application behaves when another service's endpoint misbehaves. Have your application send the requests to fakettp, and have fakettp proxy requests to the original service or return specific responses.

Sample Usage
------------

From the source or compiled binary, just set the response you want (note that you can pass in multiple headers to be returned):

```
$ go run main.go -port 5555 -code 201 -header 'Content-Type: application/json' -header 'Cache-Control: max-age=3600' -body '{"json":true}'
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