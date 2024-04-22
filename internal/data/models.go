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
	Movies          MovieModel
	ModuleInfos     ModuleInfoModel
	DepartmentInfos DepartmentInfoModel
	Permissions     PermissionModel // Add a new Permissions field.
	Users           UserModel
	Tokens          TokenModel
	UserInfos       UserInfoModel
}

// method which returns a Models struct containing the initialized MovieModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies:          MovieModel{DB: db},
		ModuleInfos:     ModuleInfoModel{DB: db},
		DepartmentInfos: DepartmentInfoModel{DB: db},
		Permissions:     PermissionModel{DB: db}, // Initialize a new PermissionModel instance.
		Users:           UserModel{DB: db},
		Tokens:          TokenModel{DB: db},
		UserInfos:       UserInfoModel{DB: db},
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

type DepartmentInfo struct {
	ID                 int    `json:"id"`
	DepartmentName     string `json:"departmentName"`
	StaffQuantity      int    `json:"staffQuantity"`
	DepartmentDirector string `json:"departmentDirector"`
	ModuleId           int    `json:"moduleId"`
}
