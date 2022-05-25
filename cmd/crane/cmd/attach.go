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
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/opencontainers/go-digest"
	specsv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

// NewCmdAttach creates a new cobra.Command for the attach subcommand.
func NewCmdAttach(options *[]crane.Option) *cobra.Command {
	return &cobra.Command{
		Use:   "attach",
		Short: "Attach a text file to an image (via reference)",
		Long:  "TODO",
		Args:  cobra.ExactArgs(4),
		RunE: func(_ *cobra.Command, args []string) error {
			fileName := args[0]
			mediaType := args[1]
			artifactType := args[2]
			src := args[3]

			// Load file
			layerContent, err := ioutil.ReadFile(fileName)
			if err != nil {
				return err
			}

			// Get the existing image manifest
			// TODO: any easier way to do this? crane.Pull() converts the mediatype...
			orig, err := crane.Manifest(src, *options...)
			if err != nil {
				return err
			}
			var tmp = struct {
				MediaType types.MediaType `json:"mediaType"`
			}{}
			err = json.Unmarshal(orig, &tmp)
			if err != nil {
				return err
			}
			origMediaType := tmp.MediaType
			origSize := int64(len(orig))
			origDigest, err := v1.NewHash(digest.FromBytes(orig).String())
			if err != nil {
				return err
			}

			// Create a new "image" with reference to the existing image
			base := mutate.Annotations(empty.Image, map[string]string{
				"org.opencontainers.artifact.type": artifactType,
			}).(v1.Image)

			base = mutate.MediaType(base, specsv1.MediaTypeImageManifest)
			base = mutate.ConfigMediaType(base, specsv1.MediaTypeImageConfig)
			base = mutate.Reference(base, &v1.Descriptor{
				MediaType: origMediaType,
				Size:      origSize,
				Digest:    origDigest,
			})

			layer := static.NewLayer(layerContent, types.MediaType(mediaType))
			img, err := mutate.Append(base, mutate.Addendum{
				Layer: layer,
			})
			if err != nil {
				return err
			}

			// TODO: any easier way to push by digest?
			dstDigest, err := img.Digest()
			if err != nil {
				return err
			}
			dstRef, err := name.ParseReference(src)
			if err != nil {
				return err
			}
			dst := fmt.Sprintf("%s@%s", dstRef.Context(), dstDigest.String())

			// Push the new "image"
			return crane.Push(img, dst, *options...)
		},
	}
}
