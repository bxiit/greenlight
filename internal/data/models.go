package data

import (
	"database/sql"
	"errors"
	"time"
)

// Define a custom ErrRecordNotFound error. We'll return this from our Get() method when
// looking up a movie that doesn't exist in our database.
var (
	ErrRecordNotFound = errors.New("record (row, entry) not found")
)

// Create a Models struct which wraps the MovieModel
// kind of enveloping
type Models struct {
	Movies      MovieModel
	ModuleInfos ModuleInfoModel
}

// method which returns a Models struct containing the initialized MovieModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		ModuleInfos: ModuleInfoModel{DB: db},
	}
}

type ModuleInfo struct {
	ID             int           `json:"id"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
	ModuleName     string        `json:"moduleName"`
	ModuleDuration time.Duration `json:"moduleDuration"`
	ExamType       string        `json:"examType"`
	Version        string        `json:"version"`
}
