package entity

type Category struct {
	BaseEntity
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Slug        string `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	Description string `gorm:"type:text" json:"description"`
	Posts       []Post `gorm:"foreignKey:CategoryID" json:"posts,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}