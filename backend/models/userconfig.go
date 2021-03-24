package models

import (
	"time"
)

// TODO
// Allow saving either 1 or 2 in layout style field
// Default layout style to 1

type UserConfig struct {
	ID          uint
	UserID      uint
	LayoutStyle int // 1 for list layout and 2 for grid layout
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
