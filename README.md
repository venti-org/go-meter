# go-meter

## build

```sh
CGO_ENABLED=0 go install github.com/venti-org/go-meter/cmd/gmeter@latest
# or local build
CGO_ENABLED=0 go build ./cmd/gmeter
```

## 

```sh
./gmeter get -u http://httpbin.org/get -c 2 -n 4
# 12 client send 12 * 10 = 120 requests
./gmeter post -u http://httpbin.org/post --bodies-path request.json -c 12 -n 10
```
