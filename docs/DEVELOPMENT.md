# Contributing

## Dependencies

This is currently a Golang project.  Other languages may be added at a later date.  See the [go.mod file](go.mod) for the current
version of Golang.

Otherwise, `make` is employed.  And of course `git`.

## Build

To simply build the `rac` binary:
```shell
make build
```

To build the binary with a `kubectl-` prefix to enable basic [kubectl plugin support](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/):
```shell
make kubectl
```

To install the `rac` binary to a directory in your execution path (default is `/usr/local/bin` but can be changed with the
`INSTALL_LOCATION` environment variable):
```shell
make install
```

And similarly, to install the `kubectl-rac` binary:
```shell
make kubectl-install
```

## Test

Currently only Golang unit tests are present:
```shell
make test
```