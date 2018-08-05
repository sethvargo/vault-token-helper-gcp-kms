// Copyright 2018 Seth Vargo.
// Copyright 2018 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sethvargo/vault-token-helper-gcp-kms/version"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudkms/v1"
)

var (
	stdin  = os.Stdin
	stderr = os.Stderr
	stdout = os.Stdout

	// cryptoKeyID is the full resource ID of the crypto key
	cryptoKeyID string

	// ctxTimeout is the operation timeout for encrypt/decrypt operations
	ctxTimeout time.Duration

	// tokenPath is the path on disk to the token
	tokenPath string

	// kms is the KMS service
	kms      *cloudkms.Service
	kmsScope = "https://www.googleapis.com/auth/cloudkms"
)

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func realMain() error {
	args := os.Args[1:]
	if len(args) < 1 {
		return fmt.Errorf("expected at least 1 argument")
	}

	var fn func() error
	switch args[0] {
	case "store":
		fn = handleStore
	case "get":
		fn = handleGet
	case "erase":
		fn = handleErase
	case "version", "-version", "--version", "-v":
		return handleVersion()
	default:
		return fmt.Errorf("invalid command %q", args[0])
	}

	cryptoKeyID = os.Getenv("VAULT_GCP_KMS_CRYPTO_KEY_ID")
	if cryptoKeyID == "" {
		return errors.New("missing VAULT_GCP_KMS_CRYPTO_KEY_ID")
	}

	if raw := os.Getenv("VAULT_GCP_KMS_TIMEOUT"); raw != "" {
		var err error
		ctxTimeout, err = time.ParseDuration(raw)
		if err != nil {
			return errors.Wrap(err, "failed to parse timeout")
		}
	} else {
		ctxTimeout = 10 * time.Second
	}

	tokenPath = os.Getenv("VAULT_GCP_KMS_ENCRYPTED_TOKEN_PATH")
	if tokenPath == "" {
		homeDir, err := homedir.Dir()
		if err != nil {
			return errors.Wrap(err, "failed to get homedir")
		}
		tokenPath = homeDir + "/.vault-gcp-kms-encrypted-token"
	}

	httpClient, err := google.DefaultClient(context.Background(), kmsScope)
	if err != nil {
		return errors.Wrap(err, "failed to create http client")
	}

	kms, err = cloudkms.New(httpClient)
	if err != nil {
		return errors.Wrap(err, "failed to create kms client")
	}
	kms.UserAgent = fmt.Sprintf("%s/%s (+%s; %s)",
		version.Name, version.Version, version.URL, runtime.Version())

	return fn()
}

func handleGet() error {
	// Get the ciphertext from disk
	ciphertext, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(err, "failed to read encrypted token from disk")
	}

	// Decrypt the ciphertext
	ctx, done := context.WithTimeout(context.Background(), ctxTimeout)
	defer done()

	resp, err := kms.Projects.Locations.KeyRings.CryptoKeys.Decrypt(cryptoKeyID, &cloudkms.DecryptRequest{
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}).
		Context(ctx).
		Fields("plaintext").
		Do()
	if err != nil {
		return errors.Wrap(err, "failed to decrypt ciphertext")
	}

	// Decode the plaintext and print the token
	plaintext, err := base64.StdEncoding.DecodeString(resp.Plaintext)
	if err != nil {
		return errors.Wrap(err, "failed to decode plaintext")
	}
	fmt.Fprintf(stdout, "%s", plaintext)

	return nil
}

func handleStore() error {
	r := bufio.NewReader(stdin)
	plaintext, err := r.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed to read value from stdin")
	}

	// Encrypt the plaintext
	ctx, done := context.WithTimeout(context.Background(), ctxTimeout)
	defer done()

	resp, err := kms.Projects.Locations.KeyRings.CryptoKeys.Encrypt(cryptoKeyID, &cloudkms.EncryptRequest{
		Plaintext: base64.StdEncoding.EncodeToString(plaintext),
	}).
		Context(ctx).
		Fields("ciphertext").
		Do()
	if err != nil {
		return errors.Wrap(err, "failed to encrypt plaintext")
	}

	// Decode the ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(resp.Ciphertext)
	if err != nil {
		return errors.Wrap(err, "failed to decode ciphertext")
	}

	// Write to file
	f, err := os.OpenFile(tokenPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to open file for writing")
	}
	defer f.Close()

	if _, err := f.Write(ciphertext); err != nil {
		return errors.Wrap(err, "failed to write contents to file")
	}

	return nil
}

func handleErase() error {
	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to remove encrypted token file")
	}
	return nil
}

func handleVersion() error {
	fmt.Fprintf(stderr, version.HumanVersion)
	return nil
}
