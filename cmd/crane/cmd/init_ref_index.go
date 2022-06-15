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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewCmdInitRefIndex creates a new cobra.Command for the attach subcommand.
func NewCmdInitRefIndex(options *[]crane.Option) *cobra.Command {
	return &cobra.Command{
		Use:   "init-ref-index",
		Short: "TODO",
		Long:  "TODO",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			src := args[0]
			ref, err := name.ParseReference(src)
			if err != nil {
				return errors.Wrap(err, "name.ParseReference()")
			}
			remoteOpts := []remote.Option{remote.WithAuthFromKeychain(authn.DefaultKeychain)}
			index, err := remote.Index(ref, remoteOpts...)
			if err != nil {
				return errors.Wrap(err, "remote.Index()")
			}
			indexManifest, err := index.IndexManifest()
			if err != nil {
				return errors.Wrap(err, "index.IndexManifest()")
			}
			if len(indexManifest.Manifests) > 0 {
				if indexManifest.Manifests[0].MediaType.IsIndex() {
					fmt.Println("Note: Image is already a reference index. Skipping.")
					return nil
				}
			}
			refIndex, err := index.RefImageIndex()
			if err != nil {
				return errors.Wrap(err, "index.RefImageIndex()")
			}
			refDigest, err := index.Digest()
			if err != nil {
				return errors.Wrap(err, "index.Digest()")
			}
			// TODO: when to push "fallback" tag vs. back to original tag
			target := fmt.Sprintf("%s:%s-%s", ref.Context().Name(), refDigest.Algorithm, refDigest.Hex)
			targetRef, err := name.ParseReference(target)
			if err != nil {
				return errors.Wrap(err, "name.ParseReference()")
			}
			err = remote.WriteIndex(targetRef, refIndex, remoteOpts...)
			if err != nil {
				return errors.Wrap(err, "remote.Write()")
			}
			return nil
		},
	}
}
