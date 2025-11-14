
.PHONY: coverage
coverage:
	go test -v -coverprofile cover.out ./...
	go tool cover -html cover.out -o cover.html

.PHONY: test
test:
	go test -v ./...

# runs the examples/testlots.go file which will process lots of images.
.PHONY: testlots
testlots:
	cd examples && go run testlots.go

.PHONY: lint
lint:
	golangci-lint run