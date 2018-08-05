// Copyright 2018 Seth Vargo.
// Copyright 2018 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package main

import (
	"fmt"
	"os"

	"github.com/sethvargo/vault-token-helper-gcp-kms/version"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "missing argument")
	}

	switch args[0] {
	case "name":
		fmt.Fprintf(os.Stdout, version.Name)
	case "version":
		fmt.Fprintf(os.Stdout, version.Version)
	default:
		fmt.Fprintf(os.Stderr, "unknown value: %s", args[0])
		os.Exit(2)
	}
}
