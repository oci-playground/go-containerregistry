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
	"io/ioutil"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewCmdAttachCatImage creates a new cobra.Command for the attach-cat-image subcommand.
func NewCmdAttachCatImage(options *[]crane.Option) *cobra.Command {
	return &cobra.Command{
		Use:   "attach-cat-image",
		Short: "TODO",
		Long:  "TODO",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			remoteOpts := []remote.Option{remote.WithAuthFromKeychain(authn.DefaultKeychain)}

			catImagePath := args[0]

			src := args[1]
			ref, err := name.ParseReference(src)
			if err != nil {
				return errors.Wrap(err, "name.ParseReference()")
			}

			// Load the image
			b, err := ioutil.ReadFile(catImagePath)
			if err != nil {
				return errors.Wrap(err, "ioutil.ReadFile()")
			}
			layer := static.NewLayer(b, "image/png") // TODO: detect mediatype

			index, err := remote.Index(ref, remoteOpts...)
			if err != nil {
				return errors.Wrap(err, "remote.Index()")
			}
			indexManifest, err := index.IndexManifest()
			if err != nil {
				return errors.Wrap(err, "index.IndexManifest()")
			}
			if len(indexManifest.Manifests) > 0 {
				if !indexManifest.Manifests[0].MediaType.IsIndex() {
					index, err = index.RefImageIndex()
					if err != nil {
						return errors.Wrap(err, "index.RefImageIndex()")
					}
					indexManifest, err = index.IndexManifest()
					if err != nil {
						return errors.Wrap(err, "index.IndexManifest()")
					}
				}
			}

			// Extract the digest of the nested index layer (this is the thing the cat image will reference)
			if len(indexManifest.Manifests) == 0 {
				return errors.New("len(indexManifest.Manifests) == 0")
			}
			nestedIndexDigest := indexManifest.Manifests[0].Digest

			// First push the blob
			err = remote.WriteLayer(ref.Context(), layer, remoteOpts...)
			if err != nil {
				return errors.Wrap(err, "remote.WriteLayer()")
			}

			// Then push an OCI manifest
			img := mutate.MediaType(empty.Image, types.OCIManifestSchema1)
			img = mutate.ConfigMediaType(img, types.OCIConfigJSON)
			img, err = mutate.AppendLayers(img, layer)
			if err != nil {
				return errors.Wrap(err, "mutate.AppendLayers()")
			}
			imgDigest, err := img.Digest()
			if err != nil {
				return errors.Wrap(err, "img.Digest()")
			}
			target := fmt.Sprintf("%s@%s", ref.Context().Name(), imgDigest.String())
			targetRef, err := name.ParseReference(target)
			if err != nil {
				return errors.Wrap(err, "name.ParseReference()")
			}
			err = remote.Write(targetRef, img)
			if err != nil {
				return errors.Wrap(err, "remote.Write()")
			}

			// Finally, append to the original index
			imageDigestStr := imgDigest.String()
			indexManifest, err = index.IndexManifest()
			if err != nil {
				return errors.Wrap(err, "index.IndexManifest()")
			}
			foundIndex := -1
			for i, v := range indexManifest.Manifests {
				if v.Digest.String() == imageDigestStr {
					foundIndex = i
				}
			}
			if foundIndex != -1 {
				return fmt.Errorf("Error: this cat image has already been attached to root object (at .manifests[%d])", foundIndex)
			}
			imgMediaType, err := img.MediaType()
			if err != nil {
				return errors.Wrap(err, "img.MediaType()")
			}
			imgSize, err := img.Size()
			if err != nil {
				return errors.Wrap(err, "img.Size()")
			}
			index = mutate.AppendManifests(index, mutate.IndexAddendum{
				Add: empty.Image,
				Descriptor: v1.Descriptor{
					MediaType: imgMediaType,
					Digest:    imgDigest,
					Size:      imgSize,
					Annotations: map[string]string{
						"org.opencontainers.reference.type":        "cat",
						"org.opencontainers.reference.digest":      nestedIndexDigest.String(),
						"org.opencontainers.reference.description": "Picture of cat",
					},
				},
			})
			err = remote.WriteIndex(ref, index, remoteOpts...)
			if err != nil {
				return errors.Wrap(err, "remote.Write()")
			}
			return nil
		},
	}
}
