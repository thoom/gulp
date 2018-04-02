# gulp

Gulp is a CLI-based HTTP client for JSON-based APIs. While it's possible to send/receive payloads other than JSON, the client
provides defaults and enhancements for JSON requests and responses.

Gulp is built around 2 concepts using JSON/YAML:
1. configuration
2. payload

## configuration

By default, the client will look for a `.gulp.yml` file in the current directory. If found, it will include the following options as part of every request. Use the `-c` argument to load a different configuration file.

### config options

* __url__: The url to use with requests. This will be overridden if the last argument in the command starts with http.

* __headers__: An map of request headers to be included in all requests. Individual headers can be overridden using the `-H` argument.

* __display__: How to display responses. If not set, only the response body will be displayed. Allowed values are `verbose` and `success-only`. These can be overridden by the `-dr`, `-ds`, and `-dv` flags. 

## payload

You can use either JSON or YAML as a payload to a posted endpoint. The advantage of using YAML is that the format is simpler than JSON and allows features like comments.

Since valid JSON is also VALID YAML, you may use either.

To submit a payload, do something like: `gulp -m POST https://ex.io/api/message < postData.yml`

# dependencies

    github.com/ghodss/yaml

# install

    go get github.com/thoom/gulp

# upgrade

    go get -u github.com/thoom/gulp


