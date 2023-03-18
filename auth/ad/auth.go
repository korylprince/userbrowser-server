package ad

import (
	"fmt"

	adauth "github.com/korylprince/go-ad-auth/v3"
	"github.com/korylprince/userbrowser-server/v3/auth"
)

// Permissions is a mapping of groups to GradeRanges
type Permissions map[string][]auth.GradeRange

// Auth represents an Active Directory authentication mechanism
type Auth struct {
	config      *adauth.Config
	permissions Permissions
}

// New returns a new *Auth with the given configuration and permissions mapping
func New(config *adauth.Config, permissions Permissions) *Auth {
	return &Auth{config: config, permissions: permissions}
}

// Authenticate authenticates the given credentials and returns the User associated with the account if successful,
// or nil if not. If an error occurs it is returned.
func (a *Auth) Authenticate(username, password string) (user *auth.User, err error) {
	var groups []string
	for group := range a.permissions {
		groups = append(groups, group)
	}

	status, entry, userGroups, err := adauth.AuthenticateExtended(a.config, username, password, []string{"displayName"}, groups)
	if err != nil {
		return nil, fmt.Errorf("Error attempting to authenticate as %s: %v", username, err)
	}

	if !status || len(userGroups) == 0 {
		return nil, nil
	}

	var gradePermissions []auth.GradeRange

	for _, group := range userGroups {
		gradePermissions = append(gradePermissions, a.permissions[group]...)
	}

	return &auth.User{
		Username:    username,
		DisplayName: entry.GetAttributeValue("displayName"),
		Permissions: gradePermissions,
	}, nil
}
