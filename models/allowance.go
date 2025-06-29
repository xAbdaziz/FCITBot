package models

import "time"

type Allowance struct {
	Date time.Time `gorm:"primaryKey;not null"`
}
