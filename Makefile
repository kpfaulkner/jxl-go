
coverage:
	go test -v -coverprofile cover.out ./...
	go tool cover -html cover.out -o cover.html

test:
	go test -v ./...

# runs the examples/testlots.go file which will process lots of images.
testlots:
	cd examples && go run testlots.go

lint:
	golangci-lint run