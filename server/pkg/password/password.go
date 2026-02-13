package password

import (
	"golang.org/x/crypto/bcrypt"
)

// Hasher 密码哈希接口
type Hasher interface {
	// Hash 哈希密码
	Hash(password string) (string, error)

	// Verify 验证密码
	Verify(hash, password string) bool
}

// BcryptHasher bcrypt实现
type BcryptHasher struct {
	cost int
}

// NewHasher 创建密码哈希器
func NewHasher() Hasher {
	return &BcryptHasher{cost: 10}
}

// Hash 哈希密码
func (b *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Verify 验证密码
func (b *BcryptHasher) Verify(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
