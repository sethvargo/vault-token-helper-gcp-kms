// Copyright 2018 Seth Vargo.
// Copyright 2018 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package version

import "fmt"

const (
	Name    = "vault-token-helper-gcp-kms"
	Version = "0.1.0"
	URL     = "https://github.com/sethvargo/vault-token-helper-gcp-kms"
)

var (
	GitCommit string

	HumanVersion = fmt.Sprintf("%s %s (%s)", Name, Version, GitCommit)
)
