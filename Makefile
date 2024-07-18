.PHONY: demo deps test build_mac build_windows build_linux init

demo:
	go run cmd/shifu/main.go run demo

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
	mv main.exe shifu.exe

build_linux: test
	GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-s -w" cmd/shifu/main.go
	mv main shifu

init:
	rm -r -f test
	go run cmd/shifu/main.go init test
	go run cmd/shifu/main.go run test
