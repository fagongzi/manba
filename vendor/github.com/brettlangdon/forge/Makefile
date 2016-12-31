test: lint go-test

lint:
	./lint.sh

go-test:
	go test

bench:
	go test -bench . -benchmem

coverage:
	out=`mktemp -t "forge-coverage"`; go test -coverprofile=$$out && go tool cover -html=$$out

deps:
	go get golang.org/x/tools/cmd/cover
	go get github.com/golang/lint/golint

.PHONY: test lint go-test bench coverage deps
