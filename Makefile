test:
	go test -count=1 -race -cover ./...
.PHONY: test

audit:
	go list -json -m all | nancy sleuth --exclude-vulnerability-file ./.nancy-ignore
.PHONY: audit

.PHONY: lint
lint:
	exit

build:
	go build ./...
.PHONY: build
