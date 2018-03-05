VAULT_ADDR?=http://127.0.0.1:8200

debug:
ifdef VAULT_TOKEN
	HUMAN_LOG=1 go run example.go
else
	VAULT_TOKEN="$(shell make vault-show-token)" HUMAN_LOG=1 go run example.go
endif

vault-create-policy:
	-@vault policy write -address=$(VAULT_ADDR) read-and-create-dataset policy.hcl 2>&1 | fgrep -v 'Success! Uploaded policy'

vault-create-token: vault-create-policy
ifndef TOKEN_CREATE_OUTPUT
TOKEN_CREATE_OUTPUT:="$(shell vault token create -address=$(VAULT_ADDR) -policy=read-and-create-dataset -period=50m -display-name=dp-dataset-exporter)"
endif

vault-show-token: vault-create-token
	@echo "$(TOKEN_CREATE_OUTPUT)" | awk '$$5=="token"{print $$6}'

vault: vault-show-token

test:
	go test -cover $(shell go list ./... | grep -v /vendor/)

.PHONY: vault debug vault-create-policy vault-create-token vault-show-token
