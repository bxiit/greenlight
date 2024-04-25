package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"github.com/bxiit/greenlight/internal/validator"
	"time"
)

type UserInfoModel struct {
	DB *sql.DB
}

type UserInfo struct {
	ID             int64     `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Name           string    `json:"name"`
	Surname        string    `json:"surname"`
	Email          string    `json:"email"`
	PasswordHashed password  `json:"-"`
	Role           string    `json:"role"`
	Activated      bool      `json:"activated"`
	Version        int       `json:"-"`
}

var AnonymousUserInfo = &UserInfo{
	ID:      0,
	Name:    "anon",
	Surname: "anon",
}

func (u *UserInfo) IsAnonymous() bool {
	return u == AnonymousUserInfo
}

func (m UserInfoModel) Insert(userInfo *UserInfo) error {
	query := `
			INSERT INTO user_info (fname, sname, email, password_hash, user_role, activated)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at, version`
	args := []interface{}{userInfo.Name, userInfo.Surname, userInfo.Email, userInfo.PasswordHashed.hash, userInfo.Role, userInfo.Activated}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&userInfo.ID, &userInfo.CreatedAt, &userInfo.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "user_info_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m UserInfoModel) Get(id int64) (*UserInfo, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
			SELECT *
			FROM user_info
			WHERE id = $1`

	var userInfo UserInfo

	err := m.DB.QueryRow(query, id).Scan(
		&userInfo.ID,
		&userInfo.CreatedAt,
		&userInfo.UpdatedAt,
		&userInfo.Name,
		&userInfo.Surname,
		&userInfo.Email,
		&userInfo.PasswordHashed.hash,
		&userInfo.Role,
		&userInfo.Activated,
		&userInfo.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &userInfo, nil
}

func (m UserInfoModel) GetByEmail(email string) (*UserInfo, error) {
	query := `
			SELECT id, created_at, updated_at, fname, sname, email, password_hash, user_role, activated, version
			FROM public.user_info
			WHERE email = $1
`
	var userInfo UserInfo
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&userInfo.ID,
		&userInfo.CreatedAt,
		&userInfo.UpdatedAt,
		&userInfo.Name,
		&userInfo.Surname,
		&userInfo.Email,
		&userInfo.PasswordHashed.hash,
		&userInfo.Role,
		&userInfo.Activated,
		&userInfo.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &userInfo, nil
}

func (m UserInfoModel) GetAll() ([]*UserInfo, error) {
	query := `SELECT id, created_at, updated_at, fname, sname, email, password_hash, user_role, activated, version FROM user_info`

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userInfos := []*UserInfo{}
	for rows.Next() {
		userInfo := &UserInfo{}
		err = rows.Scan(
			&userInfo.ID,
			&userInfo.CreatedAt,
			&userInfo.UpdatedAt,
			&userInfo.Name,
			&userInfo.Surname,
			&userInfo.Email,
			&userInfo.PasswordHashed.hash,
			&userInfo.Role,
			&userInfo.Activated,
			&userInfo.Version,
		)
		if err != nil {
			return nil, err
		}
		userInfos = append(userInfos, userInfo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userInfos, nil
}

func (m UserInfoModel) Update(userInfo *UserInfo) error {
	query := `
			UPDATE user_info
			SET updated_at = now(),
			    fname = $1, 
			    sname = $2, 
			    email = $3, 
			    version = version + 1
			WHERE id = $4 AND version = $5
			RETURNING version
`

	args := []interface{}{
		userInfo.Name,
		userInfo.Surname,
		userInfo.Email,
		userInfo.ID,
		userInfo.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&userInfo.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m UserInfoModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM user_info
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

func (m UserInfoModel) GetForToken(tokenScope, tokenPlaintext string) (*UserInfo, error) {
	// Calculate the SHA-256 hash of the plaintext token provided by the client.
	// Remember that this returns a byte *array* with length 32, not a slice.
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	query := `
			SELECT public.user_info.id, 
			       public.user_info.created_at, 
			       public.user_info.updated_at, 
			       public.user_info.fname, 
			       public.user_info.sname, 
			       public.user_info.email, 
			       public.user_info.password_hash, 
			       public.user_info.activated, 
			       public.user_info.user_role, 
			       public.user_info.version
			FROM public.user_info
			INNER JOIN user_info_tokens
			ON public.user_info.id = user_info_tokens.user_info_id
			WHERE user_info_tokens.hash = $1
			AND user_info_tokens.scope = $2
			AND user_info_tokens.expiry > $3
`
	// Create a slice containing the query arguments. Notice how we use the [:] operator
	// to get a slice containing the token hash, rather than passing in the array (which
	// is not supported by the pq driver), and that we pass the current time as the
	// value to check against the token expiry.
	args := []interface{}{tokenHash[:], tokenScope, time.Now()}
	var userInfo UserInfo
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&userInfo.ID,
		&userInfo.CreatedAt,
		&userInfo.UpdatedAt,
		&userInfo.Name,
		&userInfo.Surname,
		&userInfo.Email,
		&userInfo.PasswordHashed.hash,
		&userInfo.Activated,
		&userInfo.Role,
		&userInfo.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Return the matching user.
	return &userInfo, nil
}

func ValidateUserInfo(v *validator.Validator, userInfo *UserInfo) {
	v.Check(userInfo.Name != "", "name", "must be provided")
	v.Check(len(userInfo.Name) <= 500, "name", "must not be more than 500 bytes long")
	// Call the standalone ValidateEmail() helper.
	ValidateEmail(v, userInfo.Email)
	// If the plaintext password is not nil, call the standalone
	// ValidatePasswordPlaintext() helper.
	if userInfo.PasswordHashed.plaintext != nil {
		ValidatePasswordPlaintext(v, *userInfo.PasswordHashed.plaintext)
	}
	// If the password hash is ever nil, this will be due to a logic error in our
	// codebase (probably because we forgot to set a password for the userInfo). It's a
	// useful sanity check to include here, but it's not a problem with the data
	// provided by the client. So rather than adding an error to the validation map we
	// raise a panic instead.
	if userInfo.PasswordHashed.hash == nil {
		panic("missing password hash for userInfo")
	}
}

func (m UserInfoModel) FindNotActivatedAndExpired() ([]*UserInfo, error) {
	query := `
			SELECT u.*
			FROM user_info u
			INNER JOIN public.user_info_tokens uit on u.id = uit.user_info_id
			WHERE u.activated = false AND uit.expiry < now()
`

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*UserInfo
	for rows.Next() {
		var user UserInfo
		err = rows.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.Name, &user.Surname, &user.Email, &user.PasswordHashed.hash, &user.Role, &user.Activated, &user.Version)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserInfoModel) DeleteExpiredToken(id int64) error {
	query := `
		DELETE FROM user_info_tokens WHERE user_info_id = $1
`
	_, err := m.DB.Query(query, id)
	return err
}
