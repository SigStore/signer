/*
Copyright 2021 The Sigstore Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package payload

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
)

// Simple describes the structure of a basic container image signature payload,
// as defined at:
// https://www.redhat.com/en/blog/container-image-signing
type Simple struct {
	Critical Critical               `json:"critical"`
	Optional map[string]interface{} `json:"optional"`
}

type Critical struct {
	Identity Identity `json:"identity"`
	Image    Image    `json:"image"`
	Type     string   `json:"type"`
}

type Identity struct {
	DockerReference string `json:"docker-reference"`
}

type Image struct {
	DockerManifestDigest string `json:"docker-manifest-digest"`
}

type ImagePayload struct {
	Type   string
	Image  name.Digest
	Claims map[string]interface{}
}

func (p ImagePayload) MarshalJSON() ([]byte, error) {
	simple := Simple{
		Critical: Critical{
			Identity: Identity{
				DockerReference: p.Image.Repository.Name(),
			},
			Image: Image{
				DockerManifestDigest: p.Image.DigestStr(),
			},
			Type: p.Type,
		},
		Optional: p.Claims,
	}
	return json.Marshal(simple)
}

var _ json.Marshaler = ImagePayload{}

func (p *ImagePayload) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		// JSON "null" is a no-op by convention
		return nil
	}
	var simple Simple
	if err := json.Unmarshal(data, &simple); err != nil {
		return err
	}
	digestStr := simple.Critical.Identity.DockerReference + "@" + simple.Critical.Image.DockerManifestDigest
	digest, err := name.NewDigest(digestStr)
	if err != nil {
		return fmt.Errorf("could not parse image digest string %q: %v", digestStr, err)
	}
	p.Image = digest
	p.Claims = simple.Optional
	p.Type = simple.Critical.Type
	return nil
}

var _ json.Unmarshaler = (*ImagePayload)(nil)