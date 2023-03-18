package auth

// GradeRange represents an inclusive range of grades
type GradeRange struct {
	MinGrade int
	MaxGrade int
}

// In returns true if i is in the GradeRange
func (r GradeRange) In(i int) bool {
	return r.MinGrade <= i && i <= r.MaxGrade
}

// User represents an authenticated user
type User struct {
	Username    string
	DisplayName string
	Permissions []GradeRange
}

// Authorized returns true if the User has permissions for the give grade
func (u *User) Authorized(grade int) bool {
	for _, perm := range u.Permissions {
		if perm.In(grade) {
			return true
		}
	}

	return false
}

// Auth represents an authentication mechanism
type Auth interface {
	// Authenticate authenticates the given credentials and returns the User associated with the account if successful,
	// or nil if not. If an error occurs it is returned.
	Authenticate(username, password string) (user *User, err error)
}
