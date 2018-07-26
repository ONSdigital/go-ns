package vault

import (
	"encoding/json"
	"errors"

	vaultapi "github.com/hashicorp/vault/api"
)

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrVersionNotFound  = errors.New("version not found")
	ErrMetadataNotFound = errors.New("metadata not found")
	ErrDataNotFound     = errors.New("data not found")
	ErrVersionInvalid   = errors.New("version failed to convert to number")
)

// VaultClient Used to read and write secrets from vault
type VaultClient struct {
	client *vaultapi.Client
}

// CreateVaultClient by providing an auth token, vault address and the maximum number of retries for a request
func CreateVaultClient(token, vaultAddress string, retries int) (*VaultClient, error) {
	config := vaultapi.Config{Address: vaultAddress, MaxRetries: retries}
	client, err := vaultapi.NewClient(&config)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)
	return &VaultClient{client}, nil
}

// CreateVaultClientTLS is like the CreateVaultClient function but wraps the HTTP client with TLS
func CreateVaultClientTLS(token, vaultAddress string, retries int, cacert, cert, key string) (*VaultClient, error) {
	config := vaultapi.Config{Address: vaultAddress, MaxRetries: retries}
	config.ConfigureTLS(&vaultapi.TLSConfig{CACert: cacert, ClientCert: cert, ClientKey: key})
	client, err := vaultapi.NewClient(&config)
	if err != nil {
		return nil, err
	}

	client.SetToken(token)
	return &VaultClient{client}, nil
}

// Healthcheck determines the state of vault
func (c *VaultClient) Healthcheck() (string, error) {
	resp, err := c.client.Sys().Health()
	if err != nil {
		return "vault", err
	}

	if !resp.Initialized {
		return "vault", errors.New("vault not initialised")
	}

	return "", nil
}

// Read reads a secret from vault. If the token does not have the correct policy this returns an error;
// if the vault server is not reachable, return all the information stored about the secret.
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

// Write writes a secret to vault. Returns an error if the token does not have the correct policy or the
// vault server is not reachable. Returns nil when the operation was successful.
func (c *VaultClient) Write(path string, data map[string]interface{}) error {
	_, err := c.client.Logical().Write(path, data)
	return err
}

func (c *VaultClient) WriteKey(path, key, value string) error {
	data := make(map[string]interface{})
	data[key] = value
	return c.Write(path, data)
}

// VRead reads a versioned secret from vault - cf Read, above -
// returns the secret (map) and the version
func (c *VaultClient) VRead(path string) (map[string]interface{}, int64, error) {
	secret, err := c.Read(path)
	if err != nil {
		return nil, -1, err
	}
	if len(secret) == 0 {
		// if there is no secret and no error return a empty map
		return secret, -1, nil
	}
	metadata, ok := secret["metadata"]
	if !ok {
		return nil, -1, ErrMetadataNotFound
	}
	verObj, ok := metadata.(map[string]interface{})["version"]
	if !ok {
		return nil, -1, ErrVersionNotFound
	}
	ver, err := verObj.(json.Number).Int64()
	if err != nil {
		return nil, -1, ErrVersionInvalid
	}
	data, ok := secret["data"]
	if !ok {
		return nil, -1, ErrDataNotFound
	}
	return data.(map[string]interface{}), ver, err
}

// VReadKey - cf Read but for versioned secret - return the value of the key and the version
func (c *VaultClient) VReadKey(path, key string) (string, int64, error) {
	data, ver, err := c.VRead(path)
	if err != nil {
		return "", -1, err
	}
	val, ok := data[key]
	if !ok {
		return "", -1, ErrKeyNotFound
	}
	return val.(string), ver, nil
}

func (c *VaultClient) VWriteKey(path, key, value string) error {
	data := map[string]interface{}{
		"data": map[string]interface{}{
			key: value,
		},
	}
	return c.Write(path, data)
}
