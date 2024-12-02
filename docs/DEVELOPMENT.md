# Contributing

## Dependencies

Currently:
- `go`
- `make`
- `git`

See the [go.mod file](go.mod) for the current version of Golang used in implementing this CLI.

Implementations in other languages may occur at a later date.  The idea being implementations in other languages
may help with engagement in one upstream community versus another.

## Build

To simply build the `bac` binary:
```shell
make build
```

To build the binary with a `kubectl-` prefix to enable basic [kubectl plugin support](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/):
```shell
make kubectl
```

To install the `bac` binary to a directory in your execution path (default is `/usr/local/bin` but can be changed with the
`INSTALL_LOCATION` environment variable):
```shell
make install
```

And similarly, to install the `kubectl-bac` binary:
```shell
make kubectl-install
```

## Test

Currently only Golang unit tests are present:
```shell
make test
```