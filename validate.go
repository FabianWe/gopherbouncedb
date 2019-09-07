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

// These variables define errors returned by some of the validators.
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

// VerifiyNameExists tests if the user has a username.
// It is a UserVerifier.
func VerifiyNameExists(u *UserModel) error {
	if strings.TrimSpace(u.Username) == "" {
		return ErrEmptyUsername
	}
	return nil
}

// VerifyEmailExists tests if the user has an email.
// It is a UserVerifier.
func VerifyEmailExists(u *UserModel) error {
	if strings.TrimSpace(u.EMail) == "" {
		return ErrEmptyEmail
	}
	return nil
}

// VerifyPasswordExists tests if the password is not empty.
// It is a UserVerifier.
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

// IsEmailSyntaxValid tests if the email is syntactically correct.
// It does however not check the length of the email.
func IsEmailSyntaxValid(email string) error {
	if EmailRx.MatchString(email) {
		return nil
	}
	return ErrInvalidEmailSyntax
}

// VerifyEmailSyntax tests if the user email is syntactically.
// It does however not check the length of the email.
func VerifyEmailSyntax(u *UserModel) error {
	return IsEmailSyntaxValid(u.EMail)
}

// CheckUsernameMaxLen tests if the username is not longer than the allowed length
// (150 chars).
func CheckUsernameMaxLen(username string) error {
	if utf8.RuneCountInString(username) > 150 {
		return ErrUsernameTooLong
	}
	return nil
}

// CheckPasswordHashMaxLen tests if the password hash length is not longer than the
// allowed length (270 chars).
func CheckPasswordHashMaxLen(password string) error {
	if utf8.RuneCountInString(password) > 270 {
		return ErrPasswordTooLong
	}
	return nil
}

// CheckEmailMaxLen tests if the email length is not longer than the allowed length
// (254 chars).
func CheckEmailMaxLen(email string) error {
	if utf8.RuneCountInString(email) > 254 {
		return ErrEmailTooLong
	}
	return nil
}

// CheckFirstNameMaxLen tests if the name is not longer than the allowed length
// (50 chars).
func CheckFirstNameMaxLen(name string) error {
	if utf8.RuneCountInString(name) > 50 {
		return ErrFirstNameTooLong
	}
	return nil
}

// CheckLastNameMaxLen tests if the name is not longer than the allowed length
// (150 chars).
func CheckLastNameMaxLen(name string) error {
	if utf8.RuneCountInString(name) > 150 {
		return ErrLastNameTooLong
	}
	return nil
}

// VerifyStandardUserMaxLens tests the username, password hash, email, first name
// and last name for their max lengths and returns nil only iff all tests passed.
func VerifyStandardUserMaxLens(u *UserModel) error {
	if err := CheckUsernameMaxLen(u.Username); err != nil {
		return err
	}
	if err := CheckPasswordHashMaxLen(u.Password); err != nil {
		return err
	}
	if err := CheckEmailMaxLen(u.EMail); err != nil {
		return err
	}
	if err := CheckFirstNameMaxLen(u.FirstName); err != nil {
		return err
	}
	if err := CheckLastNameMaxLen(u.LastName); err != nil {
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
	UsernameRx = regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9]|[_.][a-zA-Z0-9])*?$`)
)

func verifyName(name string) bool {
	for _, char := range name {
		if !unicode.IsLetter(char) {
			return false
		}
	}
	return true
}

// CheckUsernameSyntax tests if the username matches the following syntax:
// First a alphabetic symbol (a-zA-Z), followed by a sequence of chars, dots
// points and numbers.
// But it is not allowed to end with an underscore or dot.
// Also after a dot or underscore no dot or underscore is allowed.
// It does however not check the length limits.
func CheckUsernameSyntax(username string) error {
	if UsernameRx.MatchString(username) {
		return nil
	}
	return ErrInvalidUsernameSyntax
}

// CheckFirstNameSyntax tests if all chars of the name are a unicode letter (class L).
func CheckFirstNameSyntax(name string) error {
	if verifyName(name) {
		return nil
	}
	return ErrInvalidFirstNameSyntax
}

// CheckLastNameSyntax tests if all chars of the name are a unicode letter (class L).
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

// PWLenVerifier is a generator for a PasswordVerifier that checks the length of
// the password.
//
// The password must have at least length minLen and at most length maxLen.
// To disable any of the checks pass -1.
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

// RuneClass is any set of runes identified by a function.
type RuneClass func(r rune) bool

// ClassCounter counts how many classes are are contained in the input string.
// That is if a class matches a single rune in s it is considered contained in
// the input.
func ClassCounter(classes []RuneClass, s string) int {

	// check class concurrently, write back to channel (either true or false)
	ch := make(chan bool)
	asRunes := []rune(s)
	for _, class := range classes {
		go func(c RuneClass) {
			fulfilled := false
			for _, char := range asRunes {
				if c(char) {
					fulfilled = true
					break
				}
			}
			ch <- fulfilled
		}(class)
	}

	numFulfilled := 0
	for i := 0; i < len(classes); i++ {
		fulfilled := <-ch
		if fulfilled {
			numFulfilled++
		}
	}
	return numFulfilled
}

// PWContainsAll is a generator that returns a PasswordVerifier.
//
// The returned verifier tests if the password contains at least one char of each class.
func PWContainsAll(classes []RuneClass) PasswordVerifier {
	n := len(classes)
	return func(pw string) bool {
		return ClassCounter(classes, pw) == n
	}
}

// LowerLetterClass tests if 'a' ≤ r ≤ 'z'.
func LowerLetterClass(r rune) bool {
	return r >= 'a' && r <= 'z'
}

// UpperLetterClass tests if 'A' ≤ r ≤ 'Z'.
func UpperLetterClass(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// DigitClass tests if the rune is a number (0 - 9).
func DigitClass(r rune) bool {
	return r >= '0' && r <= '9'
}

// SpecialCharacterClass tests if the rune is a special char from
// ~!@#$%^&*()+=_-{}[]\\|:;?/<>,
func SpecialCharacterClass(r rune) bool {
	return strings.ContainsAny(string(r), "~!@#$%^&*()+=_-{}[]\\|:;?/<>,")
}
