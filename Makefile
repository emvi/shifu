.PHONY: demo deps test build_mac build_windows build_linux init

demo:
	modd

deps:
	go get -u -t ./...
	go mod vendor

test:
	go test -cover ./pkg/cfg
	go test -cover -race $$(go list ./pkg/... | grep -v /cfg)

benchmark:
	go test -bench=. ./pkg/...

build_mac: test
	esbuild --bundle --minify --sourcemap --outfile=static/admin/static/admin.js static/admin/assets/js/main.js && \
    sass -I static/admin/assets/sass --style=compressed static/admin/assets/sass/main.scss static/admin/static/admin.css && \
    GOOS=darwin go build -a -installsuffix cgo -ldflags "-s -w" -o shifu cmd/shifu/main.go

build_windows: test
	esbuild --bundle --minify --sourcemap --outfile=static/admin/static/admin.js static/admin/assets/js/main.js && \
    sass -I static/admin/assets/sass --style=compressed static/admin/assets/sass/main.scss static/admin/static/admin.css && \
    GOOS=windows GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-s -w" -o shifu.exe cmd/shifu/main.go

build_linux: test
	esbuild --bundle --minify --sourcemap --outfile=static/admin/static/admin.js static/admin/assets/js/main.js && \
    sass -I static/admin/assets/sass --style=compressed static/admin/assets/sass/main.scss static/admin/static/admin.css && \
    GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-s -w" -o shifu cmd/shifu/main.go

build_arm64: test
	esbuild --bundle --minify --sourcemap --outfile=static/admin/static/admin.js static/admin/assets/js/main.js && \
    sass -I static/admin/assets/sass --style=compressed static/admin/assets/sass/main.scss static/admin/static/admin.css && \
	GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -ldflags "-s -w" -o shifu cmd/shifu/main.go

init:
	rm -r -f test
	go run cmd/shifu/main.go init test
	go run cmd/shifu/main.go run test
