.PHONY: demo demo_s3 deps test build_mac build_windows build_linux init

demo:
	go run cmd/shifu/main.go run demo

demo_s3:
	go run cmd/shifu/main.go run demo_s3

deps:
	go get -u -t ./...
	go mod vendor

test:
	go test -cover ./pkg/cfg
	go test -cover -race $$(go list ./pkg/... | grep -v /cfg)

benchmark:
	go test -bench=. ./pkg/...

build_mac: test
	GOOS=darwin go build -a -installsuffix cgo -ldflags "-s -w" cmd/shifu/main.go
	mv main shifu

build_windows: test
	GOOS=windows GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-s -w" -o shifu.exe cmd/shifu/main.go
	mv main.exe shifu.exe

build_linux: test
	GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-s -w" cmd/shifu/main.go
	mv main shifu

build_arm64: test
	GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -ldflags "-s -w" cmd/shifu/main.go
	mv main shifu

init:
	rm -r -f test
	go run cmd/shifu/main.go init test
	go run cmd/shifu/main.go run test
