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
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
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
			indexMediatype, err := index.MediaType()
			if err != nil {
				return err
			}
			indexDigest, err := index.Digest()
			if err != nil {
				return err
			}
			indexSize, err := index.Size()
			if err != nil {
				return err
			}

			// Create the new index
			refIndex := mutate.IndexMediaType(empty.Index, types.OCIImageIndex)
			refIndex = mutate.AppendManifests(refIndex, mutate.IndexAddendum{
				Add: empty.Image,
				Descriptor: v1.Descriptor{
					MediaType: indexMediatype,
					Digest:    indexDigest,
					Size:      indexSize,
					Annotations: map[string]string{
						"org.favorite.icecream": "mint-chocolate",
					},
				},
			})
			indexManifest, err := index.IndexManifest()
			if err != nil {
				return err
			}
			for _, manifest := range indexManifest.Manifests {
				refIndex = mutate.AppendManifests(refIndex, mutate.IndexAddendum{
					Add:        empty.Image,
					Descriptor: manifest,
				})
			}

			raw, err := refIndex.RawManifest()
			if err != nil {
				return err
			}
			fmt.Println(string(raw))

			return nil
		},
	}
}
