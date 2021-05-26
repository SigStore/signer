//
// Copyright 2021 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package signfile

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/go-openapi/swag"
	"github.com/pkg/errors"

	"github.com/sigstore/rekor/cmd/rekor-cli/app"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/sigstore/sigstore/pkg/tlog"
)

type SignedPayload struct {
	Base64Signature string
	Payload         []byte
	Cert            *x509.Certificate
	Chain           []*x509.Certificate
}

func UploadToRekor(publicKey crypto.PublicKey, signedMsg []byte, rekorURL string, certPEM []byte, payload []byte) (string, error) {
	rekorClient, err := app.GetRekorClient(rekorURL)
	if err != nil {
		return "", err
	}

	re := tlog.RekorEntry(payload, signedMsg, certPEM)
	returnVal := models.Rekord{
		APIVersion: swag.String(re.APIVersion()),
		Spec:       re.RekordObj,
	}
	params := entries.NewCreateLogEntryParams()
	params.SetProposedEntry(&returnVal)
	resp, err := rekorClient.Entries.CreateLogEntry(params)

	if err != nil {
		// If the entry already exists, we get a specific error.
		// Here, we display the proof and succeed.
		if e, ok := err.(*entries.CreateLogEntryConflict); ok {
			fmt.Printf("Signature already exists at %v\n", e.Location.String())
			return e.Location.String(), nil
		}
		return "", err
	}
	// UUID is at the end of location
	return resp.Location.String(), nil
}

func MarshalPublicKey(pub crypto.PublicKey) ([]byte, error) {
	if pub == nil {
		return nil, errors.New("empty key")
	}
	pubKey, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		panic("failed to marshall public key")
	}
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKey,
	})
	return pubBytes, nil
}
