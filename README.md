# gulp
Simplified http cli client w/ YAML-based configuration
## dependencies

    github.com/ghodss/yaml

## install

    go get github.com/thoom/gulp

## upgrade

    go get -u github.com/thoom/gulp


## config options

By default, the client will look for a `.gulp.yml` file in the current directory. If found, it will include the following options as part of every request. Use the `-c` argument to load a different configuration file.

* __url__: The url to use with requests. This will be overridden if the last argument in the command starts with http.

* __headers__: An map of request headers to be included in all requests. Individual headers can be overridden using the `-H` argument.

* __display__: How to display responses. If not set, only the response body will be displayed. Allowed values are `verbose` and `success-only`. These can be overridden by the `-dr`, `-ds`, and `-dv` flags. 
