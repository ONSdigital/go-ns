
VAULT_ADDR='http://127.0.0.1:8200'
VAULT_TOKEN=$(shell cat ~/.vault-token)
VAULT_POLICY:="$(shell vault policy write -address=$(VAULT_ADDR) read-and-create-dataset policy.hcl)"
TOKEN_INFO:="$(shell vault token create -address=$(VAULT_ADDR) -policy=read-and-create-dataset -period=50m -display-name=dp-dataset-exporter)"
APP_TOKEN:="$(shell echo $(TOKEN_INFO) | awk '{print $$6}')"
OTHER=$(VAULT_TOKEN)
unexport VAULT_TOKEN

debug:
	HUMAN_LOG=1 VAULT_TOKEN=$(APP_TOKEN) go run example.go

vault:
	@echo "$(VAULT_POLICY)"
	@echo "$(TOKEN_INFO)"
	@echo "$(APP_TOKEN)"

test:
	go test -cover $(shell go list ./... | grep -v /vendor/)

.PHONY: vault debug