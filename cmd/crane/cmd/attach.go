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
	"io/ioutil"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
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
			src := args[2]
			dst := args[3]

			// Load file
			layerContent, err := ioutil.ReadFile(fileName)
			if err != nil {
				return err
			}

			// Get the existing image manifest
			// TODO: should we just use the raw manifest?
			// vs. just grabbing size/digest/mediatype/annotations?
			orig, err := crane.Pull(src, *options...)
			if err != nil {
				return err
			}
			/*
				TODO: are the annotations new? or copied from manifest
				manifest, err := orig.Manifest()
				if err != nil {
					return err
				}
				annotations := manifest.Annotations
			*/
			origDigest, err := orig.Digest()
			if err != nil {
				return err
			}
			origSize, err := orig.Size()
			if err != nil {
				return err
			}
			origMediaType, err := orig.MediaType()
			if err != nil {
				return err
			}

			// Create a new "image" with reference to the existing image
			base := mutate.MediaType(empty.Image, specsv1.MediaTypeImageManifest)
			base = mutate.ConfigMediaType(base, specsv1.MediaTypeImageConfig)
			base = mutate.Reference(base, &v1.Descriptor{
				MediaType: origMediaType,
				Size:      origSize,
				Digest:    origDigest,
				//Annotations: annotations,
			})

			layer := static.NewLayer(layerContent, types.MediaType(mediaType))
			img, err := mutate.Append(base, mutate.Addendum{
				Layer: layer,
			})
			if err != nil {
				return err
			}

			// Push the new "image"
			return crane.Push(img, dst, *options...)
		},
	}
}
