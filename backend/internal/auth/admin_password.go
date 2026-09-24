package auth

import "golang.org/x/crypto/bcrypt"

// HashAdminPassword bcrypt-hashes an admin password for storage in
// admin_users.password_hash.
func HashAdminPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// VerifyAdminPassword reports whether password matches hash, as produced
// by HashAdminPassword.
func VerifyAdminPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
