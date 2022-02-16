# GULP ![Builds](https://github.com/thoom/gulp/actions/workflows/main.yml/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/thoom/gulp)](https://goreportcard.com/report/github.com/thoom/gulp) [![Coverage](https://sonarcloud.io/api/project_badges/measure?project=gulp&metric=coverage)](https://sonarcloud.io/summary/overall?id=gulp) [![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=gulp&metric=security_rating)](https://sonarcloud.io/summary/overall?id=gulp) [![GoDoc](https://godoc.org/github.com/thoom/gulp?status.svg)](https://godoc.org/github.com/thoom/gulp)
 
GULP is a silly acronym for **G**et **U**r**L** **P**ayload. It is an HTTP REST client written in Go. Since it's primarily meant to work with modern REST APIs, by default it expects either JSON or YAML payloads. YAML is effortlessly converted to JSON when used as a payload.

Some advantages to using YAML instead of JSON include being able to have comments and not requiring superfluous usage of curly braces and quotation marks.

For instance, a sample YAML configuration file:

```
# Some comment here...
url: https://api.github.com
headers:
  X-Example-Header: abc123def
  X-Example-Header2: ghi456jkl
flags:
  use_color: true
```

Its JSON equivalent:

```
{
  "url": "https://api.github.com",
  "headers": {
    "X-Example-Header": "abc123def",
    "X-Example-Header2": "ghi456jkl"
  },
  "flags": {
    "use_color": true
 }
}
```

GULP uses YAML/JSON for:

1. configuration
2. payload

## Installation

There are several ways to download and install the `gulp` client.

### Binary Releases

The preferred method is to download the latest binary release for your platform from the [Github Releases](https://github.com/thoom/gulp/releases) section.

### Using Docker

If you already use Docker, GULP is packaged into a very small, OS-less image (~4 Mb compressed, 7.5 Mb uncompressed). To learn more about the Docker image, see the [Github Packages](https://github.com/users/thoom/packages/container/package/gulp) section.

**Basic usage**

```
docker run --rm -it -v $PWD:/gulp ghcr.io/thoom/gulp
```

### Using Go

Finally, you can also just build it directly on your machine if you already have Go installed:

```
go get github.com/thoom/gulp
```

## Usage
Once installed, the client is easy to use without extra configuration. By default, the client makes a GET request to the endpoint.
For instance to get user _foo_'s data from the Github API:

```
gulp https://api.github.com/users/foo
```

Want more info about the request, like the request headers passed and the response headers received?

```
gulp -v https://api.github.com/users/foo
```

Imagine that you are going to be working frequently with the Github API. 
Create a configuration file (details described below) to simplify the interactions.

```
# .gulp.yml
url: https://api.github.com
```

Now you can just call:

```
gulp -v /users/foo
```

This exposes how the client builds the final URL from 2 parts: the _config.URL_ and the _Path_.

The cli format is technically in the format `gulp [FLAGS] [PATH]`. If a configuration file exists,
and it has the `url` (_config.URL_) field defined (as shown above), then it will take the _[PATH]_ from the 
cli and concatinate it with the _config.URL_. This was seen in the previous example.

If the _[PATH]_ starts with `http`, then the client will ignore the _config.URL_.

If the _[PATH]_ is empty, then the client will just use the _config.URL_ if it exists.

If both are empty, then an error is returned.

## CLI Flags

```
-H request
		Set a request header
-c configuration
		The configuration file to use (default ".gulp.yml")
-client-cert string
		If using client cert auth, the cert to use. MUST be paired with -client-cert-key flag
-client-cert-key string
		If using client cert auth, the key to use. MUST be paired with -client-cert flag
-follow-redirect
		Enables following 3XX redirects (default)
-insecure
		Disable TLS certificate checking
-m method
		The method to use: ie. HEAD, GET, POST, PUT, DELETE (default "GET")
-no-color
		Disables color output for the request
-no-redirect
		Disables following 3XX redirects
-repeat-concurrent connections
		Number of concurrent connections to use (default 1)
-repeat-times iteration
		Number of iterations to submit the request (default 1)
-ro
		Only display the response body (default)
-sco
		Only display the response code
-timeout seconds
		The number of seconds to wait before the connection times out (default 300)
-v		Display the response body along with various headers
-version
		Display the current client version
```

## Configuration

By default, the client will look for a `.gulp.yml` file in the current directory. 
If found, it will include the following options as part of every request. 
Use the `-c` argument to load a different configuration file.

### YAML Configuration Options

* __url__: The url to use with requests. 
	Setting this configuration option allows for simplified paths in the command line.
	It can be overridden if the last argument in the command starts with `http`.  

* __headers__: A map of request headers to be included in all requests. 
	Individual headers can be overridden using the `-H` argument.

* __display__: How to display responses.
	If not set, only the response body will be displayed.
	Allowed values are `verbose` and `status-code-only`.
	These can be overridden by the `-ro`, `-sco`, and `-v` cli flags. 

* __timeout__: How long to wait for a response from the remote server.
	Defaults to 300 seconds. Can be overridden by the `-timeout` cli argument.

* __client_auth__: The file and key to use with client cert requests.
  * __cert__: The PEM-encoded file to use as cert
  * __key__:  The PEM-encoded file to use as private key

* __flags__: Options that are enabled by default and can be disabled:
  * __follow_redirects__: Follow `3XX` HTTP redirects. 
	Can be disabled with the `-no-redirect` flag.
  
  * __use_color__: Colorize verbose responses. 
	Can be disabled with the `-no-color` flag.
  
  * __verify_tls__: Verify SSL/TLS certificates. 
	Can be disabled with the `-insecure` flag.

## POST Payload

Since GULP prefers JSON/YAML payloads _(Note: YAML is converted to JSON automatically)_, using either is easy. 

### To post a payload of JSON or YAML

The command:

```
gulp -m POST https://api.ex.io/message < postData.yml
```

OR

```
cat postData.yml | gulp -m POST https://api.ex.io/message
```

### To post a payload other than JSON/YAML

The command is slightly more complicated since a content-type must be passed. The command:

```
gulp -m POST -H "Content-Type: image/jpeg" https://api.ex.io/photo < me.jpg
```

OR 

```
cat me.jpg | gulp -m POST -H "Content-Type: image/jpeg" https://api.ex.io/photo
```

## Load Testing

There are 2 command line flags that can be used as a poor-man's load testing/throttling service:

 * __-repeat-times__: The number of times to submit a request.
 
 * __-repeat-concurrent__: The number of concurrent connections to use to submit the request.

 For example, if you ran `gulp -repeat-times 100 -repeat-concurrent 10 /some/api`, 
 the CLI would make 100 total requests with a concurrency of 10 calls at a time (so it would average about 10 calls per thread).
 
 ## Library Dependencies

	github.com/fatih/color
	github.com/ghodss/yaml
	github.com/stretchr/testify (tests only)

    
