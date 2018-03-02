package vault

import (
	"errors"

	vaultapi "github.com/hashicorp/vault/api"
)

// VaultClient Used to read and write secrets from vault
type VaultClient struct {
	token  string
	client *vaultapi.Client
}

// CreateVaultClient by providing a auth token, vault address and the maxium number of retries for a request
func CreateVaultClient(token, vaultAddress string, retries int) (*VaultClient, error) {
	config := vaultapi.Config{Address: vaultAddress, MaxRetries: retries}
	client, err := vaultapi.NewClient(&config)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)
	return &VaultClient{token, client}, nil
}

// Read a secret from vault. If the token does not have the correct policy this shall return an error or if the
// vault server is not reachable. This method shall return all the information store about the secret
func (c *VaultClient) Read(path string) (map[string]interface{}, error) {
	secret, err := c.client.Logical().Read(path)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		// If there is no secret and no error return a empty map.
		return make(map[string]interface{}), nil
	}
	return secret.Data, err
}

// ReadKey from vault. Like read but only return a single value from the secret
func (c *VaultClient) ReadKey(path, key string) (string, error) {
	data, err := c.Read(path)
	if err != nil {
		return "", err
	}
	val, ok := data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return val.(string), nil
}

// Write a secret to vault. If the token does not have the correct policy this shall return an error or if the
// vault server is not reachable. If nil is returned the operation was successful
func (c *VaultClient) Write(path string, data map[string]interface{}) error {
	_, err := c.client.Logical().Write(path, data)
	return err
}

func (c *VaultClient) WriteKey(path, key, value string) error {
	data := make(map[string]interface{})
	data[key] = value
	return c.Write(path, data)
}
