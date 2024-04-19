package data

import (
	"database/sql"
	"errors"
	"log"
)

type DepartmentInfoModel struct {
	DB *sql.DB
}

func (m DepartmentInfoModel) Insert(departmentInfo *DepartmentInfo) error {
	query := `INSERT INTO department_info(department_name, staff_quantity, department_director, module_id)
			VALUES ($1, $2, $3, $4)
			RETURNING id`

	log.Println("inserting to database")

	args := []interface{}{departmentInfo.DepartmentName, departmentInfo.StaffQuantity, departmentInfo.DepartmentDirector, departmentInfo.ModuleId}
	return m.DB.QueryRow(query, args...).Scan(&departmentInfo.ID)
}

func (m DepartmentInfoModel) Get(id int64) (*DepartmentInfo, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
			SELECT * FROM department_info
			WHERE id = $1`
	var departmentInfo DepartmentInfo

	err := m.DB.QueryRow(query, id).Scan(
		&departmentInfo.ID,
		&departmentInfo.DepartmentName,
		&departmentInfo.StaffQuantity,
		&departmentInfo.DepartmentDirector,
		&departmentInfo.ModuleId,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &departmentInfo, nil
}
