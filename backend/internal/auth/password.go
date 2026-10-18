package auth

import "golang.org/x/crypto/bcrypt"

// bcryptCost is the cost factor required by RN-1.
const bcryptCost = 12

// dummyHash is a valid bcrypt hash (cost 12) used to keep the response time of
// a failed login constant whether or not the email exists. It is not a secret:
// it is never matched against a real password.
const dummyHash = "$2a$12$IAphA4RGDVa8iephOKeTjO0or0WIyxrFGZ671yHPq50V4QHUNy9IO"

// HashPassword hashes a plain password with bcrypt (cost 12).
func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// CheckPassword reports whether plain matches the stored bcrypt hash.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// DummyCheck performs a bcrypt comparison against a fixed hash so that an
// unknown email costs the same as a wrong password.
func DummyCheck(plain string) {
	_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(plain))
}
