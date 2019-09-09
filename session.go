// Copyright 2019 Fabian Wenzelmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gopherbouncedb

import (
	"crypto/rand"
	"time"
	"io"
	"encoding/base64"
)

const (
	SessionKeyBytes = 29
)

func GenSessionKey() (string, error) {
	randBytes := make([]byte, SessionKeyBytes)
	if _, genErr := io.ReadFull(rand.Reader, randBytes); genErr != nil {
		return "", genErr
	}
	return base64.RawURLEncoding.EncodeToString(randBytes), nil
}

type SessionEntry struct {
	User UserID
	Key string
	ExpireDate time.Time
}

func (s *SessionEntry) Copy() *SessionEntry {
	return &SessionEntry{
		User: s.User,
		Key: s.Key,
		ExpireDate: s.ExpireDate,
	}
}
