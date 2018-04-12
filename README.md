# gulp [![Build Status](https://travis-ci.org/thoom/gulp.svg?branch=master)](https://travis-ci.org/thoom/gulp)

Gulp is a CLI-based HTTP client for JSON-based APIs. While it's possible to send/receive payloads other than JSON, the client
provides defaults and enhancements for JSON requests and responses.

Gulp is built around 2 concepts using JSON/YAML:
1. configuration
2. payload

## configuration

By default, the client will look for a `.gulp.yml` file in the current directory. If found, it will include the following options as part of every request. Use the `-c` argument to load a different configuration file.

### config options

* __url__: The url to use with requests. Setting this configuration option allows for simplified paths in the command line. It can be overridden if the last argument in the command starts with `http`.  

  For instance, if you have the url configuration `https://api.ex.io`, then if you want to get the current user (ie. _https://api.ex.io/user/me_), the command line call could be `gulp /user/me` instead of `gulp https://api.ex.io/user/me`.

* __headers__: A map of request headers to be included in all requests. Individual headers can be overridden using the `-H` argument.

* __display__: How to display responses. If not set, only the response body will be displayed. Allowed values are `verbose` and `status-code-only`. These can be overridden by the `-ro`, `-sco`, and `-I` flags. 

* __flags__: Options that can be turned on or off:
  * __use_color__: Whether or not to colorize verbose responses. Enabled by default.

  * __verify_tls__: Whether or not to check TLS certificates. Enabled by default. Can be overridden by the `-k` flag.

## payload

You can use either JSON or YAML as a payload to a posted endpoint. Some advantages to using YAML instead of JSON include being able to have comments and not requiring superfluous usage of curly braces and quotation marks.

The command to post data: `gulp -m POST https://api.ex.io/message < postData.yml`

## load testing

There are 2 command line flags that can be used as a poor-man's load testing/throttling service:

 * __-repeat-times__: The number of times to submit a request.

 * __-repeat-concurrent__: The number of concurrent connections to use to submit the request.

 For example, if you ran `gulp -repeat-times 100 -repeat-concurrent 10 /some/api`, the CLI would make 10 concurrent requests 10 times in a row.  

# dependencies

    github.com/fatih/color
    github.com/ghodss/yaml
    github.com/stretchr/testify (tests only)
    
# install

    go get github.com/thoom/gulp

# upgrade

    go get -u github.com/thoom/gulp


