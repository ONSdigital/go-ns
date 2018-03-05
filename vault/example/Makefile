
VAULT_ADDR='http://127.0.0.1:8200'

# The following variables are used to generate a vault token for the app. The reason for declaring variables, is that
# its difficult to move the token code in a Makefile action. Doing so makes the Makefile more difficult to
# read and starts introduction if/else statements.
VAULT_POLICY:="$(shell vault policy write -address=$(VAULT_ADDR) read-and-create-dataset policy.hcl)"
TOKEN_INFO:="$(shell vault token create -address=$(VAULT_ADDR) -policy=read-and-create-dataset -period=50m -display-name=dp-dataset-exporter)"
APP_TOKEN:="$(shell echo $(TOKEN_INFO) | awk '{print $$6}')"

debug:
	HUMAN_LOG=1 VAULT_TOKEN=$(APP_TOKEN) go run example.go

vault:
	@echo "$(VAULT_POLICY)"
	@echo "$(TOKEN_INFO)"
	@echo "$(APP_TOKEN)"

test:
	go test -cover $(shell go list ./... | grep -v /vendor/)

.PHONY: vault debug