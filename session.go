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
	// SessionKeyBytes is the length of session random keys in bytes.
	// Session keys are encoded in base64 which results in strings of length 39.
	SessionKeyBytes = 29
)

// GenSessionKey returns a new cryptographically secure random key.
// It uses 29 random bytes (as defined in SessionKeyBytes). The random bytes
// are base64 encoded which results in strings of length 39.
func GenSessionKey() (string, error) {
	randBytes := make([]byte, SessionKeyBytes)
	if _, genErr := io.ReadFull(rand.Reader, randBytes); genErr != nil {
		return "", genErr
	}
	return base64.RawURLEncoding.EncodeToString(randBytes), nil
}

// SessionEntry is an entry for a session to be stored in a database.
// It describes the user this session belongs to (by id) and a unique cryptographically
// secure random key.
// The ExpireDate describes how long the session is considered valid.
type SessionEntry struct {
	Key string
	User UserID
	ExpireDate time.Time
}

// NewSessionWithKey returns a new SessionEntry and creates automatically a new
// session key.
// If an error is returned the session should not be used.
func NewSessionWithKey(user UserID, expireDate time.Time) (*SessionEntry, error) {
	key, keyErr := GenSessionKey()
	if keyErr != nil {
		return nil, keyErr
	}
	return &SessionEntry{
		User: user,
		Key: key,
		ExpireDate: expireDate,
	}, nil
}

// SessionEntry returns a copy of another session entry.
func (s *SessionEntry) Copy() *SessionEntry {
	return &SessionEntry{
		Key: s.Key,
		User: s.User,
		ExpireDate: s.ExpireDate,
	}
}

// IsValid returns true iff the session is considered valid.
// This function should be used to check if a session is valid.
// One databases it's generally a good idea not to iterate over all session and test
// whether it's still valid.
// So the definition of a valid session is that the reference date is before the
// ExpireDate of the session.
func (s *SessionEntry) IsValid(referenceDate time.Time) bool {
	return referenceDate.Before(s.ExpireDate)
}
