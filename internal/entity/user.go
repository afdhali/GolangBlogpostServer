package entity

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleAdmin      UserRole = "admin"
	RoleUser       UserRole = "user"
)

type User struct {
	BaseEntity
	Username string   `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email    string   `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password string   `gorm:"type:varchar(255);not null" json:"-"`
	FullName string   `gorm:"type:varchar(100)" json:"full_name"`
	Role     UserRole `gorm:"type:varchar(20);not null;default:'user'" json:"role"`
	IsActive bool     `gorm:"default:true" json:"is_active"`
	Avatar 	 string   `gorm:"type:varchar(255)" json:"avatar,omitempty"` 
}

func (User) TableName() string {
	return "users"
}

func (u *User) HashPassword(pass string) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPass)
	return nil
}

func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}

func (u *User) IsSuperAdmin() bool {
    return u.Role == RoleSuperAdmin
}

func (u *User) IsAdmin() bool {
    return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

func (u *User) CanManageUser(targetUserID uuid.UUID) bool {
    if u.IsSuperAdmin() {
        return true
    }
    return u.ID == targetUserID
}

// func (u *User) CanManagePost(post *Post) bool {
// 	if u.IsAdmin() {
// 		return true
// 	}
// 	return u.ID == post.AuthorID
// }

func (u *User) CanManageComment(comment *Comment) bool {
	return u.ID == comment.UserID
}

func (u *User) CanPublishPost(post *Post) bool {
	return u.IsAdmin()
}

func (u *User) CanManageCategory(category *Category) bool {
	return u.IsAdmin()
}