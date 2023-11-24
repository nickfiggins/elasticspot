GO=go
GOCOVER=$(GO) tool cover
GOTEST=${GO} test
COVERAGE_DIR=.
MODULE=github.com/nickfiggins/elasticspot
.PHONY: test lint coverage

test:
	$(GOTEST) -v -coverprofile=$(COVERAGE_DIR)/coverage.out $(MODULE)
	$(GOCOVER) -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html

coverage:
	open $(COVERAGE_DIR)/coverage.html

lint:
	golangci-lint run --disable unused --disable deadcode