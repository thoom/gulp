# gulp [![Build Status](https://travis-ci.org/thoom/gulp.svg?branch=master)](https://travis-ci.org/thoom/gulp) [![Go Report Card](https://goreportcard.com/badge/github.com/thoom/gulp)](https://goreportcard.com/report/github.com/thoom/gulp) [![Coverage Status](https://coveralls.io/repos/github/thoom/gulp/badge.svg?branch=master)](https://coveralls.io/github/thoom/gulp?branch=master)

Gulp is a YAML-oriented HTTP CLI client for JSON APIs. When interacting with an API, Gulp accepts
either JSON or YAML. Since JSON is a subset of the YAML specification, YAML payloads are effortlessly
converted to JSON when submitting to the API.

Some advantages to using YAML instead of JSON include being able to have comments 
and not requiring superfluous usage of curly braces and quotation marks.

For instance, a sample YAML configuration file:

	# Some comment here...
	url: https://api.github.com
	headers:
	  X-Example-Header: abc123def
	  X-Example-Header2: ghi456jkl
	flags:
	  use_color: true

Its JSON equivalent:

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

Gulp uses YAML/JSON for:

1. configuration
2. payload

## Installation

There are several ways to download and install the `gulp` client.

### Using Go

	go get github.com/thoom/gulp

### Using Docker

	docker run --rm -it -v $PWD:/gulp thoom/gulp

### Releases

Download the appropriate binary from the [Github releases](https://github.com/thoom/gulp/releases) section.

## Usage
Once installed, the client is easy to use without extra configuration. 
For instance to get user _foo_'s data from the Github API:

	gulp https://api.github.com/users/foo

Want more info about the request, like the request headers passed and the response headers received?

	gulp -v https://api.github.com/users/foo

Imagine that you are going to be working frequently with the Github API. 
Create a configuration file (details described below) to simplify the interactions.

	# .gulp.yml
	url: https://api.github.com

Now you can just call:

	gulp -v /users/foo

This exposes how the client builds the final URL from 2 parts: the _config.URL_ and the _Path_.

The cli format is technically in the format `gulp [FLAGS] [PATH]`. If a configuration file exists,
and it has the `url` (_config.URL_) field defined (as seen above), then it will take the _[PATH]_ from the 
cli and concatinate it with the _config.URL_. This was seen in the previous example.

If the _[PATH]_ starts with `http`, then the client will ignore the _config.URL_.

If the _[PATH]_ is empty, then the client will just use the _config.URL_.

If both are empty, then an error is returned.

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

* __flags__: Options that are enabled by default and can be disabled:
  * __follow_redirects__: Whether or not to follow `3XX` HTTP redirects. 
	Can be disabled with the `-no-redirect` flag.
  
  * __use_color__: Whether or not to colorize verbose responses. 
	Can be disabled with the `-no-color` flag.
  
  * __verify_tls__: Whether or not to check TLS certificates. 
	Can be disabled with the `-k` flag.

## POST Payload

You can use either JSON or YAML as a payload to a posted endpoint.
The command to post data: `gulp -m POST https://api.ex.io/message < postData.yml`

## Load Testing

There are 2 command line flags that can be used as a poor-man's load testing/throttling service:

 * __-repeat-times__: The number of times to submit a request.
 
 * __-repeat-concurrent__: The number of concurrent connections to use to submit the request.

 For example, if you ran `gulp -repeat-times 100 -repeat-concurrent 10 /some/api`, 
 the CLI would make 10 concurrent requests 10 times in a row.
 
 ## Library Dependencies

	github.com/fatih/color
	github.com/ghodss/yaml
	github.com/stretchr/testify (tests only)

    
