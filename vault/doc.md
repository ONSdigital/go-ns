Vault Client
============

## Setting up vault
Using brew type `brew install vault` or the latest binaries can be downloaded at https://www.vaultproject.io/downloads.html

## Running vault
To simplify running vault hashicorp has included a development mode for vault. Here is a list of things to remember when running in this mode;
* HTTPS is disabled
* An in-memory store is used (restarting vault will erase all data, including policies)
* When using the CLI tools make sure to export the following environment variable `VAULT_ADDR='http://127.0.0.1:8200'`
* When starting in development mode the vault is already in an `unsealed` state
* The first 10 log lines in development mode print the `seal` and `root-token` keys
* Once the vault server is started the `root-token` is written to `~/.vault-token` (using `vault login 'token'` will overwrite this)

To run in development mode run `vault server -dev`

An example of log output for the root token looks like this;

```
You may need to set the following environment variable:

    $ export VAULT_ADDR='http://127.0.0.1:8200'

The unseal key and root token are displayed below in case you want to
seal/unseal the Vault or re-authenticate.

Unseal Key: 4oHp2u/l6w2uPiZ/jqOe6EHTN6op6Xj7il9MUZCY7Ic=
Root Token: cb0a36cf-12c5-6b8c-87c9-63f990661f2e
```

## Running the vault client example

To run the example code, make sure vault is started in development mode (`vault server -dev`). Then using the 
Makefile run `make debug`

If the example ran successfully you should see the following output;

```
2018-03-02 12:55:16.278977145 +0000 GMT m=+0.003972743 debug: successfully  written and read a key from vault
  -> address: http://127.0.0.1:8200
  -> token: c7598b3b-d9f5-465e-8e10-0bb7ee86f423
  -> pk-key: 098980474948463874535354
```

Behind the scenes some additional operations were carried out to create a unique vault token for this app. All
this is done in the Makefile. The following operations are carried out;

* Require the root_token
* Create a policy to allow the app to `read` and `create` in `secret/shared/datasets/*`
* Using the newly created policy generates a token for the application
* Inject the generated token to the app from an environment variable

The root token could be used for the app, but the app will have the permission to carry out
any operations in vault. This does not allow the policy to be tested when developing the app

The policy used for this app can be found at `policy.hcl`