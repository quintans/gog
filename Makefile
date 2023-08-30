# Run tests
.PHONY: test
test:
	go test -count=1 -v ./...

.PHONY: install
install:
	go install .
