package models

import "time"

type GroupsNotes struct {
	ID          uint   `gorm:"primaryKey"`
	GroupID     string `gorm:"index;not null"`
	NoteName    string `gorm:"not null"`
	NoteContent string `gorm:"not null"`
	CreatedAt   time.Time
}
