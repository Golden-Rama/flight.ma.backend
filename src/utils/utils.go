package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash for the given password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// VerifyPassword verifies if the given password matches the stored hash.
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// MD5 Hash
func SetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func Base64StrEncode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Base64StrDecode(encodedString string) string {
	var decodedByte, _ = base64.StdEncoding.DecodeString(encodedString)
	return string(decodedByte)
}

// Slugify generates a slug from string
func Slugify(s string) string {
	s = strings.ToLower(s)
	// replace non-alphanumeric with dash
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	// trim dashes
	s = strings.Trim(s, "-")
	return s
}

// FirstLetterToUpper capitalizes the first letter
func FirstLetterToUpper(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// FormatTimeID formats time to Indo timezone string
func FormatTimeID(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err == nil {
		t = t.In(loc)
	}
	return t.Format("2006-01-02 15:04:05")
}
