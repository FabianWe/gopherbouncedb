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
	"strings"
	"regexp"
	"unicode/utf8"
	"errors"
	"unicode"
)

// UserVerifier is a function that takes a user and returns an error if a given
// criteria isn't matched.
// For example we can check if username / email / password are given.
//
// Note that this performs validation on a user model, thus it tests the password hash
// and cannot be used to verify if a password is valid according to some other
// criteria (for example minimum length).
// A clear text password should be checked before, see PasswordVerifier.
// They're also used to verify certain length properties for the database.
type UserVerifier func(u *UserModel) error

var (
	ErrEmptyUsername = errors.New("no username given")
	ErrEmptyEmail = errors.New("no EMail given")
	ErrEmptyPassword = errors.New("no password set")
	ErrInvalidEmailSyntax = errors.New("invalid syntax in email")
	ErrUsernameTooLong = errors.New("username is longer than 150 characters")
	ErrPasswordTooLong = errors.New("password is longer than 270 characters")
	ErrEmailTooLong = errors.New("email is longer than 254 characters")
	ErrFirstNameTooLong = errors.New("first name is longer than 50 characters")
	ErrLastNameTooLong = errors.New("last name is longer than 150 characters")
	ErrInvalidUsernameSyntax = errors.New("invalid username syntax")
	ErrInvalidFirstNameSyntax = errors.New("invalid first name")
	ErrInvalidLastNameSyntax = errors.New("invalid last name")
)

func VerifiyNameExists(u *UserModel) error {
	if strings.TrimSpace(u.Username) == "" {
		return ErrEmptyUsername
	}
	return nil
}

func VerifyEmailExists(u *UserModel) error {
	if strings.TrimSpace(u.EMail) == "" {
		return ErrEmptyEmail
	}
	return nil
}

func VerifyPasswordExists(u *UserModel) error {
	if strings.TrimSpace(u.Password) == "" {
		return ErrEmptyPassword
	}
	return nil
}

var (
	// EmailRegexp is used to verify that an email is valid.
	// It is the python version taken from https://emailregex.com/
	EmailRx = regexp.MustCompile(`(^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$)`)
)

func IsEmailValid(email string) error {
	if EmailRx.MatchString(email) {
		return nil
	}
	return ErrInvalidEmailSyntax
}

func VerifyEmailSyntax(u *UserModel) error {
	return IsEmailValid(u.EMail)
}

func CheckUsernameLength(username string) error {
	if utf8.RuneCountInString(username)> 150 {
		return ErrUsernameTooLong
	}
	return nil
}

func CheckPasswordHashLength(password string) error {
	if utf8.RuneCountInString(password) > 270 {
		return ErrPasswordTooLong
	}
	return nil
}

func CheckEmailLength(email string) error {
	if utf8.RuneCountInString(email) > 254 {
		return ErrEmailTooLong
	}
	return nil
}

func CheckFirstNameLength(name string) error {
	if utf8.RuneCountInString(name) > 50 {
		return ErrFirstNameTooLong
	}
	return nil
}

func CheckLastNameLength(name string) error {
	if utf8.RuneCountInString(name) > 150 {
		return ErrLastNameTooLong
	}
	return nil
}

func VerifyStandardUserLengths(u *UserModel) error {
	if err := CheckUsernameLength(u.Username); err != nil {
		return err
	}
	if err := CheckPasswordHashLength(u.Password); err != nil {
		return err
	}
	if err := CheckEmailLength(u.EMail); err != nil {
		return err
	}
	if err := CheckFirstNameLength(u.FirstName); err != nil {
		return err
	}
	if err := CheckLastNameLength(u.LastName); err != nil {
		return err
	}
	return nil
}

var (
	// UsernameSyntaxRx tries to implement the following syntax for user names:
	// First a alphabetic symbol (a-zA-Z), followed by a sequence of chars, dots
	// points and numbers.
	// But it is not allowed to end with an underscore or dot.
	// Also after a dot or underscore no dot or underscore is allowed.
	// It does however not check the length limits.
	UsernameSyntaxRx = regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9]|[_.][a-zA-Z0-9])*?$`)
)

func verifyName(name string) bool {
	for _, char := range name {
		if !unicode.IsLetter(char) {
			return false
		}
	}
	return true
}

func CheckUsernameSyntax(username string) error {
	if UsernameSyntaxRx.MatchString(username) {
		return nil
	}
	return ErrInvalidUsernameSyntax
}

func CheckFirstNameSyntax(name string) error {
	if verifyName(name) {
		return nil
	}
	return ErrInvalidFirstNameSyntax
}

func CheckLastNameSyntax(name string) error {
	if verifyName(name) {
		return nil
	}
	return ErrInvalidLastNameSyntax
}

// PasswordVerifier is any function that checks if a given password meets certain
// criteria, for example min length or contains at least one character from a certain
// range.
type PasswordVerifier func(pw string) bool

func PWLenVerifier(minLen, maxLen int) PasswordVerifier {
	return func(pw string) bool {
		pwLen := utf8.RuneCountInString(pw)
		if minLen >= 0 && pwLen < minLen {
			return false
		}
		if maxLen >= 0 && pwLen > maxLen {
			return false
		}
		return true
	}
}

type RuneClass func(r rune) bool

func ClassCounter(classes []RuneClass, s string) int {
	classCounter := make(map[int]struct{}, len(classes))
	for _, char := range s {
		for i, class := range classes {
			if class(char) {
				classCounter[i] = struct{}{}
			}
		}
	}
	return len(classCounter)
}

func LowerLetterClass(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func UpperLetterClass(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func DigitClass(r rune) bool {
	return r >= '0' && r <= '9'
}

func SpecialCharacterClass(r rune) bool {
	return strings.ContainsAny(string(r), "~!@#$%^&*()+=_-{}[]\\|:;?/<>,")
}

func LetterClass(r rune) bool {
	return LowerLetterClass(r) || UpperLetterClass(r)
}

func PWContainsVerifier(classes []RuneClass) PasswordVerifier {
	return func(pw string) bool {
		return false
	}
}
