package main

import (
	"os"

	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/vault"
)

func main() {

	log.Namespace = "vault-example"
	devAddress := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")

	client, err := vault.CreateVaultClient(token, devAddress, 3)

	// In production no tokens should be logged
	logData := log.Data{"address": devAddress, "token": token}
	log.Debug("Created vault client", logData)

	if err != nil {
		log.ErrorC("failed to connect to vault", err, logData)
	}

	err = client.WriteKey("secret/shared/datasets/CPIH-0000", "PK-Key", "098980474948463874535354")

	if err != nil {
		log.ErrorC("failed to write to vault", err, logData)
	}

	PKKey, err := client.ReadKey("secret/shared/datasets/CPIH-0000", "PK-Key")

	if err != nil {
		log.ErrorC("failed to read  PK Key from vault", err, logData)
	}
	logData["pk-key"] = PKKey
	log.Debug("successfully  written and read a key from vault", logData)
}
