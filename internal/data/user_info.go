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

//	func (u UserInfoModel) Insert(userInfo *UserInfo) error {
//		query := `INSERT INTO user_info(fname, sname, email, password_hash, user_role, activated)
//				VALUES($1, $2, $3, $4, $5, $6)`
//
//		//args := []any{userInfo.Name, userInfo.Surname, userInfo.Email, userInfo.PasswordHashed, userInfo.Role, userInfo.Activated}
//		_, err := u.DB.Exec(query, &userInfo.Name, &userInfo.Surname, &userInfo.Email, &userInfo.PasswordHashed, &userInfo.Role, &userInfo.Activated)
//		return err
//	}
//
//	func (u UserInfoModel) Get(userInfo *UserInfo) error {
//		query := `SELECT * FROM user_info WHERE id = $1`
//
//		row := u.DB.QueryRow(query, userInfo.ID)
//		err := row.Scan(&userInfo.ID, &userInfo.Name, &userInfo.Surname, &userInfo.Email, &userInfo.PasswordHashed, &userInfo.Role, &userInfo.Activated)
//		return err
//	}
//
//	func (u UserInfoModel) Update(userInfo *UserInfo) error {
//		query := `UPDATE user_info SET fname = $1, sname = $2, email = $3, password_hash = $4, user_role = $5, activated = $6 WHERE id = $7`
//
//		_, err := u.DB.Exec(query, &userInfo.Name, &userInfo.Surname, &userInfo.Email, &userInfo.PasswordHashed, &userInfo.Role, &userInfo.Activated, &userInfo.ID)
//		return err
//	}
//
//	func (u UserInfoModel) Delete(userInfo *UserInfo) error {
//		query := `DELETE FROM user_info WHERE id = $1`
//
//		_, err := u.DB.Exec(query, userInfo.ID)
//		return err
//	}
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
	//args := []any{userInfo.Name, userInfo.Surname, userInfo.Email, userInfo.PasswordHashed.hash, "user", userInfo.Activated}
	args := []interface{}{userInfo.Name, userInfo.Surname, userInfo.Email, userInfo.PasswordHashed.hash, userInfo.Role, userInfo.Activated}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// If the table already contains a record with this email address, then when we try
	// to perform the insert there will be a violation of the UNIQUE "users_email_key"
	// constraint that we set up in the previous chapter. We check for this error
	// specifically, and return custom ErrDuplicateEmail error instead.
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

// Retrieve the User details from the database based on the user's email address.
// Because we have a UNIQUE constraint on the email column, this SQL query will only
// return one record (or none at all, in which case we return a ErrRecordNotFound error).
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

// Update the details for a specific user. Notice that we check against the version
// field to help prevent any race conditions during the request cycle, just like we did
// when updating a movie. And we also check for a violation of the "users_email_key"
// constraint when performing the update, just like we did when inserting the user
// record originally.
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
	// Set up the SQL query.
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
	// Execute the query, scanning the return values into a User struct. If no matching
	// record is found we return an ErrRecordNotFound error.
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
