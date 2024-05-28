.PHONY: deps test build_mac build_windows build_linux demo

deps:
	go get -u -t ./...
	go mod vendor

test:
	go test -cover ./pkg/...

build_mac: test
	GOOS=darwin go build -a -installsuffix cgo -ldflags "-s -w" cmd/shifu/main.go
	mv main shifu

build_windows: test
	GOOS=windows GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-s -w" -o shifu.exe cmd/shifu/main.go
	mv main shifu

build_linux: test
	GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-s -w" cmd/shifu/main.go
	mv main shifu

demo:
	go run cmd/shifu/main.go run demo

init:
	rm -r -f test
	go run cmd/shifu/main.go init test
	go run cmd/shifu/main.go run test
