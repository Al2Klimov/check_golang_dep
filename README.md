## About

The check plugin **check\_golang\_dep** monitors
whether any dependencies of a [Golang][golang] [package][go-pkg]
have been updated compared to [Gopkg.lock].

## Demonstration

1. `$ docker run -itp 8080:80 grandmaster/check_golang_dep`
2. Open http://localhost:8080 and navigate to any service

## Requirements

* a \*nix-like OS
* [go]
* [git]
* [dep]

## Usage

check\_golang\_dep takes two positional CLI arguments
and no environment variables:

```
$ ./check_golang_dep GO_PACKAGE CACHE_DIR
```

GO\_PACKAGE is the Golang package to monitor the dependencies of.

CACHE\_DIR is a directory a check\_golang\_dep service
uses exclusively (per host and service!) for caching. It either...

* already exists and is read- and writable by check\_golang\_dep or
* could be created by check\_golang\_dep via e.g. `mkdir -p CACHE_DIR`
  without sudo(8) or similar.

### Legal info

To print the legal info, execute the plugin in a terminal:

```
$ ./check_golang_dep
```

In this case the program will always terminate with exit status 3 ("unknown")
without actually checking anything.

### Testing

If you want to actually execute a check inside a terminal,
you have to connect the standard output of the plugin to anything
other than a terminal – e.g. the standard input of another process:

```
$ ./check_golang_dep github.com/Al2Klimov/check_golang_dep "$(mktemp -d)" |cat
```

In this case the exit code is likely to be the cat's one.
This can be worked around like this:

```
bash $ set -o pipefail
bash $ ./check_golang_dep github.com/Al2Klimov/check_golang_dep "$(mktemp -d)" |cat
```

### Actual monitoring

Just integrate the plugin into the monitoring tool of your choice
like any other check plugin. (Consult that tool's manual on how to do that.)
It should work with any monitoring tool
supporting the [Nagio$ check plugin API].

#### Icinga 2

This repository ships the [check command definition]
as well as a [service template] and [host example] for [Icinga 2].

[golang]: https://golang.org/
[go-pkg]: https://golang.org/ref/spec#Packages
[Gopkg.lock]: https://golang.github.io/dep/docs/Gopkg.lock.html
[go]: https://golang.org/cmd/go/
[git]: https://git-scm.com/
[dep]: https://golang.github.io/dep/
[Nagio$ check plugin API]: https://nagios-plugins.org/doc/guidelines.html#AEN78
[check command definition]: ./icinga2/check_golang_dep.conf
[service template]: ./icinga2/check_golang_dep-service.conf
[host example]: ./icinga2/check_golang_dep-host.conf
[Icinga 2]: https://www.icinga.com/docs/icinga2/latest/doc/01-about/
