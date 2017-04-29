# RT
RT is a lightweight autotests runtime using [Docker](http://docker.com) to isolate test cases.

## Building

1) Install [Golang](https://golang.org/doc/install)
2) Setup `$GOPATH` [properly](https://github.com/golang/go/wiki/GOPATH)
3) Install [govendor](https://github.com/kardianos/govendor): 
```
$ go get -u github.com/kardianos/govendor
```
4) Get source:
```
$ go get -d github.com/aerokube/rt
```
5) Go to project directory:
```
$ cd $GOPATH/src/github.com/aerokube/rt
```
6) Checkout dependencies:
```
$ govendor sync
```
7) Build source:
```
$ go build
```
8) Run Selenoid:
```
$ ./rt --help
```
9) To build Docker container type:
```
$ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build
$ docker build -t rt:latest .
```
