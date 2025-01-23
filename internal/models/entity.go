package models

import (
	"fmt"
	"github.com/lucsky/cuid"
	"gorm.io/gorm"
	"time"
)

type Entity struct {
	Id          string         `json:"id" gorm:"primaryKey"`
	ParentId    *string        `json:"parent_id" gorm:"index"` // Allow null for top-level entities
	Parent      *Entity        `gorm:"foreignKey:ParentId"`    // Self-referencing relationship
	Children    []*Entity      `json:"children" gorm:"foreignKey:ParentId"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Images      []string       `json:"images" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

////////////////////////////////////////////////
// DB Hook methods
////////////////////////////////////////////////

func (e *Entity) BeforeCreate(tx *gorm.DB) (err error) {
	if e.Id == "" {
		e.Id = cuid.New()
	}
	return nil
}

func (e *Entity) BeforeDelete(tx *gorm.DB) (err error) {
	var count int64

	// Check if there are any child entities
	if err := tx.Model(&Entity{}).Where("parent_id = ?", e.Id).Count(&count).Error; err != nil {
		return fmt.Errorf("error checking for child entities: %w", err)
	}

	// If there are children, prevent deletion
	if count > 0 {
		return fmt.Errorf("cannot delete entity with ID %s because it has child entities", e.Id)
	}
	return nil
}

////////////////////////////////////////////////
// Constructors
////////////////////////////////////////////////

type NewEntityOption func(entity *Entity)

func NewEntity(
	opts ...NewEntityOption,
) *Entity {
	entity := &Entity{}
	for _, o := range opts {
		o(entity)
	}
	return entity
}

func EntityWithId(id string) NewEntityOption {
	return func(e *Entity) {
		e.Id = id
	}
}

func EntityWithName(name string) NewEntityOption {
	return func(e *Entity) {
		e.Name = name
	}
}

func EntityWithDescription(desc string) NewEntityOption {
	return func(e *Entity) {
		e.Description = desc
	}
}

func EntityWithParentId(id string) NewEntityOption {
	return func(e *Entity) {
		if id == "" {
			e.ParentId = nil
			return
		}
		e.ParentId = &id
	}
}

func EntityWithImages(imageUrls []string) NewEntityOption {
	return func(e *Entity) {
		e.Images = imageUrls
	}
}
