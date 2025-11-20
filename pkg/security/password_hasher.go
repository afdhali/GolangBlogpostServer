package security

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrHashingPassword   = errors.New("failed to hash password")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrPasswordTooShort  = errors.New("password too short")
	ErrPasswordTooLong   = errors.New("password too long")
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
	ValidatePassword(password string) error
}

type passwordHasher struct {
	cost      int
	minLength int
	maxLength int
}

// NewPasswordHasher creates a new password hasher with bcrypt
func NewPasswordHasher(cost, minLength, maxLength int) PasswordHasher {
	// Default cost if not provided
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	
	// Default min/max length
	if minLength == 0 {
		minLength = 6
	}
	if maxLength == 0 {
		maxLength = 72 // bcrypt max length
	}

	return &passwordHasher{
		cost:      cost,
		minLength: minLength,
		maxLength: maxLength,
	}
}

// Hash generates bcrypt hash from password
func (p *passwordHasher) Hash(password string) (string, error) {
	// Validate password first
	if err := p.ValidatePassword(password); err != nil {
		return "", err
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", ErrHashingPassword
	}

	return string(hashedBytes), nil
}

// Verify compares password with hash
func (p *passwordHasher) Verify(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return err
	}
	return nil
}

// ValidatePassword checks password requirements
func (p *passwordHasher) ValidatePassword(password string) error {
	if len(password) < p.minLength {
		return ErrPasswordTooShort
	}
	if len(password) > p.maxLength {
		return ErrPasswordTooLong
	}
	return nil
}

// DefaultPasswordHasher returns password hasher with default settings
// Cost: 10, MinLength: 6, MaxLength: 72
func DefaultPasswordHasher() PasswordHasher {
	return NewPasswordHasher(10, 6, 72)
}