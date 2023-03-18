package db

// User represents a student user
type User struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Grade     int    `json:"grade"`
}

// DB represents a user database
type DB interface {
	// Get returns the user with the given username, nil if the user doesn't exist, or an error if one occurred
	Get(username string) (*User, error)

	// List returns a list of all Users from the database or an error if one occurred
	List() ([]*User, error)

	// ResetPassword sets a newly generated password for the user and returns it, or an error if one occurred
	ResetPassword(username string) (string, error)
}
