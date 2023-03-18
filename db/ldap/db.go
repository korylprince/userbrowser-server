package ldap

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"sort"

	"github.com/go-ldap/ldap/v3"
	adauth "github.com/korylprince/go-ad-auth/v3"
	"github.com/korylprince/securetoken"
	"github.com/korylprince/userbrowser-server/v3/db"
)

var gradeRegexp = regexp.MustCompile("^CN=.*?,OU=(.*?)(?: Grade)?,.*$")

var gradeMap = map[string]int{
	"Pre-K":        -1,
	"Kindergarten": 0,
	"1st":          1,
	"2nd":          2,
	"3rd":          3,
	"4th":          4,
	"5th":          5,
	"6th":          6,
	"7th":          7,
	"8th":          8,
	"9th":          9,
	"10th":         10,
	"11th":         11,
	"12th":         12,
}

// DB represents a connection to an Active Directory server
type DB struct {
	config   *adauth.Config
	bindUser string
	bindPass string
	key      []byte
	debug    bool
}

// New returns a new *DB with the given parameters
func New(config *adauth.Config, username, password, key string, debug bool) *DB {
	return &DB{config: config, bindUser: username, bindPass: password, key: []byte(key), debug: debug}
}

// Bind returns a bound connection to an Active Directory server
func (d *DB) Bind() (*adauth.Conn, error) {
	conn, err := d.config.Connect()
	if err != nil {
		return nil, err
	}

	status, err := conn.Bind(d.bindUser, d.bindPass)
	if err != nil {
		return nil, err
	}

	if !status {
		return nil, fmt.Errorf("Invalid bind credentials for user %s", d.bindUser)
	}

	return conn, nil
}

// Get returns the user with the given username, nil if the user doesn't exist, or an error if one occurred
func (d *DB) Get(username string) (*db.User, error) {
	conn, err := d.Bind()
	if err != nil {
		return nil, fmt.Errorf("Error binding to server: %v", err)
	}
	defer conn.Conn.Close()

	entry, err := conn.GetAttributes("sAMAccountName", username, []string{"sn", "givenname", "sAMAccountName", "adminDescription"})
	if err != nil {
		return nil, fmt.Errorf("Error searching for user: %v", err)
	}

	if entry == nil {
		return nil, nil
	}

	var (
		grade int
		ok    bool
	)
	if match := gradeRegexp.FindStringSubmatch(entry.DN); len(match) == 2 {
		if grade, ok = gradeMap[match[1]]; !ok {
			return nil, fmt.Errorf("Unknown grade: %s", match[1])
		}
	} else {
		return nil, fmt.Errorf("Unknown grade for dn: %s", entry.DN)
	}

	pass, err := securetoken.DecryptToken(entry.GetRawAttributeValue("adminDescription"), d.key, 0)
	if err != nil {
		pass = []byte("")
		if d.debug {
			log.Printf("Unable to decrypt password for user %s: %v\n", username, err)
		}
	}

	user := &db.User{
		FirstName: entry.GetAttributeValue("givenName"),
		LastName:  entry.GetAttributeValue("sn"),
		Username:  entry.GetAttributeValue("sAMAccountName"),
		Password:  string(pass),
		Grade:     grade,
	}

	return user, nil
}

// List returns a list of all Users from the database or an error if one occurred
func (d *DB) List() ([]*db.User, error) {
	conn, err := d.Bind()
	if err != nil {
		return nil, fmt.Errorf("Error binding to server: %v", err)
	}
	defer conn.Conn.Close()

	request := ldap.NewSearchRequest(
		conn.Config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.DerefAlways,
		0,
		0,
		false,
		"(&(objectCategory=Person)(employeeID=s*)(!(UserAccountControl:1.2.840.113556.1.4.803:=2)))",
		[]string{"sn", "givenname", "sAMAccountName", "adminDescription"},
		nil,
	)

	result, err := conn.Conn.SearchWithPaging(request, 1000)
	if err != nil {
		return nil, fmt.Errorf("Error searching: %v", err)
	}

	var users []*db.User

	for _, entry := range result.Entries {
		var (
			grade int
			ok    bool
		)
		if match := gradeRegexp.FindStringSubmatch(entry.DN); len(match) == 2 {
			if grade, ok = gradeMap[match[1]]; !ok {
				log.Println("WARNING: Unknown grade for user", entry.DN)
				continue
			}
		} else {
			log.Println("WARNING: Unknown grade for user", entry.DN)
			continue
		}

		pass, err := securetoken.DecryptToken(entry.GetRawAttributeValue("adminDescription"), d.key, 0)
		if err != nil {
			pass = []byte("")
			if d.debug {
				log.Printf("Unable to decrypt password for user %s: %v\n", entry.GetAttributeValue("sAMAccountName"), err)
			}
		}

		users = append(users, &db.User{
			FirstName: entry.GetAttributeValue("givenName"),
			LastName:  entry.GetAttributeValue("sn"),
			Username:  entry.GetAttributeValue("sAMAccountName"),
			Password:  string(pass),
			Grade:     grade,
		})

	}

	sort.Slice(users, func(i, j int) bool {
		if users[i].Grade == users[j].Grade {
			if users[i].LastName == users[j].LastName {
				return users[i].FirstName < users[j].FirstName
			}
			return users[i].LastName < users[j].LastName
		}
		return users[i].Grade < users[j].Grade
	})

	return users, nil
}

// ResetPassword sets a newly generated password for the user and returns it, or an error if one occurred
func (d *DB) ResetPassword(username string) (string, error) {
	conn, err := d.Bind()
	if err != nil {
		return "", fmt.Errorf("Error binding to server: %v", err)
	}
	defer conn.Conn.Close()

	entry, err := conn.GetAttributes("sAMAccountName", username, nil)
	if err != nil {
		return "", fmt.Errorf("Error searching username %s: %v", username, err)
	}

	r, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", fmt.Errorf("Error getting random value: %v", err)
	}

	pass := fmt.Sprintf("Bullard%04d", r)

	token, err := securetoken.NewToken([]byte(pass), d.key)
	if err != nil {
		return "", fmt.Errorf("Error generating token: %v", err)
	}

	req := ldap.NewModifyRequest(entry.DN, nil)
	req.Replace("adminDescription", []string{string(token)})
	if err = conn.Conn.Modify(req); err != nil {
		return "", fmt.Errorf("Error updating token: %v", err)
	}

	err = conn.ModifyDNPassword(entry.DN, pass)
	if err != nil {
		return "", fmt.Errorf("Error modifying password: %v", err)
	}

	return pass, nil
}
