# Google Cloud Platform KMS Vault Token Helper

The Google Cloud Platform KMS Vault token helper leverages
[Google Cloud KMS][kms] for encrypting and decrypting Vault tokens stored
locally on disk.

The [default token helper][token-helper] stores tokens at `~/.vault-token`. On shared systems,
systems with tighter security requirements, or for users wishing to add an
additional layer of protection, this may not be desirable. The GCP KMS token
helper is a drop-in replacement that encrypts the tokens with GCP KMS before
persisting to disk. This has the added benefit of adding an additional layer of
authentication - the proper IAM access to the GCP KMS key.

## Installation

1. Download and install the binary from [GitHub Releases][releases]. You can
also compile it yourself, but that is not covered here.

1. Put the binary somewhere on disk, like `~/.vault.d/token-helpers`:

    ```sh
    $ mv vault-token-helper ~/.vault.d/token-helpers/vault-token-helper-gcp-kms
    ```

1. Create a Vault configuration file at `~/.vault` with the contents:

    ```hcl
    token_helper = "/Users/<your username>/.vault.d/token-helpers/vault-token-helper-gcp-kms"
    ```

    Be sure to replace `<your username>` with your username. The value must be
    an absolute path (you cannot use a relative path).

    The local CLI will automatically use this configuration.

## Usage

1. Set the `VAULT_GCP_KMS_CRYPTO_KEY_ID` environment variable to the **full
resource ID** of the KMS key. You can get this value when [creating a KMS
key][kms-create-keys].

    ```text
    export VAULT_GCP_KMS_CRYPTO_KEY_ID=projects/my-project/locations/my-location/keyRings/my-keyring/cryptoKeys/my-crypto-key
    ```

1. Use Vault normally. Commands like `vault login` will automatically delegate
to the helper, which will encrypt and decrypt the tokens behind the scenes.

### Configuration

- `VAULT_GCP_KMS_CRYPTO_KEY_ID` - the **full resource ID** of the Google Cloud
  KMS crypto key ID to use for encryption and decryption. You can get this value
  when [creating a KMS key][kms-create-keys].

- `VAULT_GCP_KMS_ENCRYPTED_TOKEN_PATH` - the path on disk where the encrypted
  token will be written. The default value is
  `$HOME/.vault-gcp-kms-encrypted-token`

- `VAULT_GCP_KMS_TIMEOUT` - the maximum time to wait for responses from KMS to
  encrypt/decrypt values. This is specified as a Golang duration value like 5s
  (5 seconds) or 10m (10 minutes). The default value is 10 seconds.

### Google Cloud Authentication

The Google Cloud KMS Vault Token Helper uses the official Google Cloud
Golang SDK. This means it supports the common ways of
[providing credentials to Google Cloud][cloud-creds].

1. The environment variable `GOOGLE_APPLICATION_CREDENTIALS`. This is specified
as the **path** to a Google Cloud credentials file, typically for a service
account. If this environment variable is present, the resulting credentials are
used. If the credentials are invalid, an error is returned.

1. Default instance credentials. When no environment variable is present, the
default service account credentials are used.

For more information on service accounts, please see the
[Google Cloud Service Accounts documentation][service-accounts].

To use this token helper, the service account must have the following minimum
scope(s):

```text
https://www.googleapis.com/auth/cloudkms
```

Additionally, the service account must have the following minimum role(s):

```text
roles/cloudkms.cryptoKeyEncrypterDecrypter
```

## FAQ

**How is this different than Vault's built-in transit secrets engine?**<br>
Great question. Vault's built-in transit secrets engine requires a Vault token
to operate, thus creating a chicken-and-egg problem. By leverage a trusted
third-party encryption system like Google Cloud KMS, not only do we avoid this
chicken-and-egg problem, but we also provide an added layer of authentication
requirement - IAM access to Google Cloud KMS - to decrypt the token stored on
disk.

[kms]: https://cloud.google.com/kms
[kms-create-keys]: https://cloud.google.com/kms/docs/creating-keys
[releases]: https://github.com/sethvargo/vault-token-helper-gcp-kms/releases
[service-accounts]: https://cloud.google.com/compute/docs/access/service-accounts
[token-helper]: https://www.vaultproject.io/docs/commands/token-helper.html
