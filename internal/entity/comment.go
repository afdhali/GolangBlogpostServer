package entity

import "github.com/google/uuid"

type Comment struct {
    BaseEntity
    Content  string     `gorm:"type:text;not null" json:"content"`
    PostID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"post_id"`
    Post     *Post      `gorm:"foreignKey:PostID" json:"post,omitempty"`
    UserID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
    User     *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
    ParentID *uuid.UUID `gorm:"type:uuid;index" json:"parent_id,omitempty"`
    Parent   *Comment   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
    Replies  []Comment  `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

func (Comment) TableName() string {
    return "comments"
}

func (c *Comment) IsReply() bool {
    return c.ParentID != nil
}