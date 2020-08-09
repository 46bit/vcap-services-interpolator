.PHONY: build clean test

APP_ROOT ?= $(PWD)

build: bin/vcap-services-interpolator

bin/vcap-services-interpolator: clean
	go build -mod=vendor -o ./bin/vcap-services-interpolator .

clean:
	rm -f bin/*

test:
	go test -mod=vendor $$(go list github.com/alphagov/paas-prometheus-endpoints/pkg/...)
	go test -mod=vendor ./...
