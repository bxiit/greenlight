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
	ModuleInfos     ModuleInfoRepo
	DepartmentInfos DepartmentInfoModel
	Permissions     PermissionRepo // Add a new Permissions field.
	Users           UserModel
	Tokens          TokenRepo
	UserInfos       UserInfoRepo
}

// method which returns a Models struct containing the initialized MovieModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies:          MovieModel{DB: db},
		ModuleInfos:     ModuleInfoRepo{DB: db},
		DepartmentInfos: DepartmentInfoModel{DB: db},
		Permissions:     PermissionRepo{DB: db}, // Initialize a new PermissionRepo instance.
		Users:           UserModel{DB: db},
		Tokens:          TokenRepo{DB: db},
		UserInfos:       UserInfoRepo{DB: db},
	}
}

type ModuleInfo struct {
	ID             int           `json:"id" gorm:"primaryKey"`
	CreatedAt      time.Time     `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt      time.Time     `json:"updatedAt" gorm:"column:updated_at"`
	ModuleName     string        `json:"moduleName" gorm:"column:module_name"`
	ModuleDuration time.Duration `json:"moduleDuration" gorm:"column:module_duration"`
	ExamType       string        `json:"examType" gorm:"column:exam_type"`
	Version        string        `json:"version" gorm:"column:version"`
}

type DepartmentInfo struct {
	ID                 int    `json:"id"`
	DepartmentName     string `json:"departmentName"`
	StaffQuantity      int    `json:"staffQuantity"`
	DepartmentDirector string `json:"departmentDirector"`
	ModuleId           int    `json:"moduleId"`
}
