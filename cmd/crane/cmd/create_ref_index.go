// Copyright 2022 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

// NewCmdCreateRefIndex creates a new cobra.Command for the attach subcommand.
func NewCmdCreateRefIndex(options *[]crane.Option) *cobra.Command {
	return &cobra.Command{
		Use:   "create-ref-index",
		Short: "TODO",
		Long:  "TODO",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			src := args[0]
			ref, err := name.ParseReference(src)
			if err != nil {
				return err
			}

			remoteOpts := []remote.Option{
				remote.WithAuthFromKeychain(authn.DefaultKeychain),
			}

			index, err := remote.Index(ref, remoteOpts...)
			if err != nil {
				return err
			}

			d, err := index.Digest()
			if err != nil {
				return err
			}
			refIndexTag := fmt.Sprintf("%s-%s", d.Algorithm, d.Hex)

			fmt.Printf("Ref index tag: %s\n", refIndexTag)

			m, err := index.IndexManifest()

			for i, layer := range m.Manifests {
				fmt.Printf("[%d] %s %s\n", i, layer.MediaType, layer.Digest)
			}
			return nil
		},
	}
}
