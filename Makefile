# Run tests
.PHONY: test
test:
	go test -race -count=1 -v ./...

.PHONY: install
install:
	go install .
