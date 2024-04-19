package data

import (
	"database/sql"
	"errors"
	"log"
)

type ModuleInfoModel struct {
	DB *sql.DB
}

// Insert method for creating a record to the moduleInfos table.
func (m ModuleInfoModel) Insert(moduleInfo *ModuleInfo) error {
	query := `
		INSERT INTO module_info(created_at, module_name, module_duration, exam_type)
		VALUES (now(), $1, $2, $3)
		RETURNING id, created_at, version`

	log.Println("inserting to database")

	//return m.DB.QueryRow(query, &moduleInfo.Title, &moduleInfo.Year, &moduleInfo.Runtime, pq.Array(&moduleInfo.Genres)).Scan(&moduleInfo.ID, &moduleInfo.CreatedAt, &moduleInfo.Version)
	return m.DB.QueryRow(
		query,
		&moduleInfo.ModuleName,
		&moduleInfo.ModuleDuration,
		&moduleInfo.ExamType,
	).Scan(
		&moduleInfo.ID,
		&moduleInfo.CreatedAt,
		&moduleInfo.Version,
	)
}

// Get method for fetching a specific record from the moduleInfos table.
func (m ModuleInfoModel) Get(id int64) (*ModuleInfo, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT *
		FROM module_info
		WHERE id = $1`

	var moduleInfo ModuleInfo

	err := m.DB.QueryRow(query, id).Scan(
		&moduleInfo.ID,
		&moduleInfo.CreatedAt,
		&moduleInfo.UpdatedAt,
		&moduleInfo.ModuleName,
		&moduleInfo.ModuleDuration,
		&moduleInfo.ExamType,
		&moduleInfo.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &moduleInfo, nil
}

func (m ModuleInfoModel) GetLatestFifty() ([]*ModuleInfo, error) {
	query := `SELECT * FROM module_info ORDER BY id DESC LIMIT 50`

	rows, err := m.DB.Query(query)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	var moduleInfos []*ModuleInfo
	for rows.Next() {
		moduleInfo := &ModuleInfo{}
		err = rows.Scan(
			&moduleInfo.ID,
			&moduleInfo.CreatedAt,
			&moduleInfo.UpdatedAt,
			&moduleInfo.ModuleName,
			&moduleInfo.ModuleDuration,
			&moduleInfo.ExamType,
			&moduleInfo.Version,
		)

		if err != nil {
			return nil, err
		}

		moduleInfos = append(moduleInfos, moduleInfo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return moduleInfos, nil
}

// Update method for updating a specific record in the moduleInfos table.
func (m ModuleInfoModel) Update(moduleInfo *ModuleInfo) error {
	query := `UPDATE module_info
			  SET updated_at = now(),
			      module_name = $1,
			      module_duration = $2,
			      exam_type = $3,
			      version = version + 1
			      WHERE id = $4
			      RETURNING version`

	args := []interface{}{
		moduleInfo.ModuleName,
		moduleInfo.ModuleDuration,
		moduleInfo.ExamType,
		moduleInfo.ID,
	}

	return m.DB.QueryRow(query, args...).Scan(&moduleInfo.Version)
}

// Delete method for deleting a specific record from the moduleInfos table.
func (m ModuleInfoModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM module_info
		WHERE id = $1`

	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
