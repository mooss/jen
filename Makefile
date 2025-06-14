.PHONY: nothing test lint

nothing:
	@echo -n 'jen '
	@git rev-parse HEAD

test:
	@$(MAKE) go-test bash-test

go-test:
	go test ./go/...

bash-test:
	./tests/context_tests.bash

lint:
	golangci-lint run ./go/...\
	  --max-issues-per-linter 0\
	  --max-same-issues 0\
	  --allow-parallel-runners\
	  --sort-results

install:
	go install go/ai/jenai.go
