package image

import (
	"errors"
	"mime/multipart"
	"strings"
)

var (
    ErrInvalidImageType = errors.New("invalid image type")
    ErrImageTooLarge    = errors.New("image size too large")
)

type Validator struct {
    maxSize       int64    // in bytes
    allowedTypes  []string
}

func NewValidator(maxSizeMB int, allowedTypes []string) *Validator {
    return &Validator{
        maxSize:      int64(maxSizeMB * 1024 * 1024),
        allowedTypes: allowedTypes,
    }
}

func (v *Validator) Validate(header *multipart.FileHeader) error {
    // Check file size
    if header.Size > v.maxSize {
        return ErrImageTooLarge
    }

    // Check content type
    contentType := header.Header.Get("Content-Type")
    if !v.isAllowedType(contentType) {
        return ErrInvalidImageType
    }

    return nil
}

func (v *Validator) isAllowedType(contentType string) bool {
    for _, allowed := range v.allowedTypes {
        if strings.Contains(contentType, allowed) {
            return true
        }
    }
    return false
}

// Default validators
func DefaultAvatarValidator() *Validator {
    return NewValidator(5, []string{"image/jpeg", "image/jpg", "image/png", "image/webp"})
}

func DefaultImageValidator() *Validator {
    return NewValidator(10, []string{"image/jpeg", "image/jpg", "image/png", "image/webp"})
}