# Shell to use with Make
SHELL := /bin/bash

# Export targets not associated with files.
.PHONY: pkg deps fmt test citest clean protobuf

# Use godep to collect dependencies.
deps:
	@echo "Fetching dependencies"
	-godep restore

# Format the Go source code
fmt:
	@echo "Formatting the source"
	-gofmt -w .

# Target for simple testing on the command line
test:
	ginkgo -r -v

# Target for testing in continuous integration
citest:
	ginkgo -r -v --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --compilers=2

# Clean build files
clean:
	@echo "Cleaning up the project source."
	-go clean
	-find . -name "*.coverprofile" -print0 | xargs -0 rm -rf
	-rm -rf site
	-rm -rf _bin
	-rm -rf _build

# Compile protocol buffers
protobuf:
	@echo "Compiling protocol buffers"
	@protoc -I rpc/ rpc/*.proto --go_out=plugins=grpc:rpc
