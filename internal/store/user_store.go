package store

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "golang.org/x/crypto/bcrypt"
)

// types declaration
type password struct {
	plaintText *string
	hash       []byte
}

// * password hashing func to be exported for use in user validations
func ( p *password) Set(plainTextPass string)error {
	hash,err := bcrypt.GenerateFromPassword([]byte(plainTextPass),12) // takes what to hash, hashingCostUnits
	if err != nil {
		return err
	}
	p.plaintText = &plainTextPass
	p.hash = hash
	return nil
} 

func (p *password) Matches (plainTextPass string) (bool,error) {
	err := bcrypt.CompareHashAndPassword(p.hash,[]byte(plainTextPass)) // passing in hashed pass n normal inputted
	if err != nil {
		switch {
		case errors.Is(err,bcrypt.ErrMismatchedHashAndPassword) :
			return false,nil
			default :
			return false,nil
		}
	}
	return true,nil
}

type User struct { // LOGGED IN USER
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash password  `json:"-"`
	Bio          string    `json:"bio"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

//* Determines which user is coming -- Auth purpose
var AnonymousUser = &User{} //* adding empty User type struct saved to this variable
func (u *User) IsAnonymousUser() bool {
	return u == AnonymousUser // checking if incoming client user matches the anony type -- when struct is empty
}

type PostgresUserStore struct {
	db *sql.DB // * so changes stay persistent across the db N server to the client
}

// ? -  function that returns instance of the type Struct 
// - creates instance of struct so whoever will references to it, the connection is shared among all meths
func NewPostUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{
		db : db,
	}
}

// interface --> let us connect methods to parent type to access these
type UserStore interface {
	CreateUser(*User) error
	GetUserByUsername(username string) (*User,error)
	UpdateUser(*User) error
	GetUserToken(scope string,tokenPlainText string) (*User, error)
 }

//! CREATEUSER METHOD -  directly access type PUsrStore
func ( s *PostgresUserStore) CreateUser(user *User) error {
	query := `
  INSERT INTO users (username, email, password_hash, bio)
  VALUES ($1, $2, $3, $4)
  RETURNING id, created_at, updated_at
  `

	err := s.db.QueryRow(query, user.Username, user.Email, user.PasswordHash.hash, user.Bio).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresUserStore) GetUserByUsername(username string) (*User, error) {
	user := &User{
		PasswordHash: password{},
	}

	query := `
  SELECT id, username, email, password_hash, bio, created_at, updated_at
  FROM users
  WHERE username = $1
  `

	err := s.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash.hash,
		&user.Bio,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *PostgresUserStore) UpdateUser(user *User) error {
	query := `
  UPDATE users
  SET username = $1, email = $2, bio = $3, updated_at = CURRENT_TIMESTAMP
  WHERE id = $4
  RETURNING updated_at
  `

	result, err := s.db.Exec(query, user.Username, user.Email, user.Bio, user.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// !auth tokenzation
func (s *PostgresUserStore) GetUserToken(scope string,plaintextpassword string) (*User,error) {
	tokenHash := sha256.Sum256([]byte(plaintextpassword)) //* get hashed pass using sha256 salt

	query := `
	 Select u.id, u.username, u.email,u.password_hash, u.bio, u.created_at, u.updated_at 
	 from users u
	 INNER JOIN tokens t
	 ON
	 t.user_id = u.id
	 WHERE
	 t.hash=$1 AND t.scope=$2 AND t.expiry > $3
	`
// intializing instance of User struct but with these fields only
	user := &User{
		PasswordHash: password{},
	}

	err := s.db.QueryRow(query,tokenHash[:],scope,time.Now()).Scan(
		// scaaning values from this query --> accessing n implementing those
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash.hash,
		&user.Bio,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows{
		// if got nothing from the query
		return nil,nil
	}
	if err != nil {
		return nil,err
	}

	return user,nil
}