package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/kelseyhightower/envconfig"
	adauth "github.com/korylprince/go-ad-auth/v3"
	"github.com/korylprince/userbrowser-server/v3/auth"
)

// parsePermissions parses the format "{Group Name}:{min-grade}<>{max-grade};{min-grade}<>{max-grade};...,..."
func parsePermissions(str string) (map[string][]auth.GradeRange, error) {
	permissions := make(map[string][]auth.GradeRange)
	for _, group := range strings.Split(str, ",") {
		splits := strings.Split(group, ":")
		if len(splits) != 2 {
			return nil, fmt.Errorf("Unable to parse group: %s", group)
		}

		groupName := strings.TrimSpace(splits[0])
		ranges := strings.TrimSpace(splits[1])

		var gradeRanges []auth.GradeRange

		for _, r := range strings.Split(ranges, ";") {

			grades := strings.Split(strings.TrimSpace(r), "<>")

			if len(grades) != 2 {
				return nil, fmt.Errorf("Unable to parse grade range: %s", r)
			}

			minGrade, err := strconv.Atoi(strings.TrimSpace(grades[0]))
			if err != nil {
				return nil, fmt.Errorf("Unable to parse minimum grade: %s: %v", grades[0], err)
			}

			maxGrade, err := strconv.Atoi(strings.TrimSpace(grades[1]))
			if err != nil {
				return nil, fmt.Errorf("Unable to parse maximum grade: %s: %v", grades[1], err)
			}

			gradeRanges = append(gradeRanges, auth.GradeRange{MinGrade: minGrade, MaxGrade: maxGrade})
		}

		permissions[groupName] = gradeRanges
	}

	if len(permissions) == 0 {
		return nil, fmt.Errorf("No permissions found: %s", str)
	}
	return permissions, nil
}

// Config represents options given in the environment
type Config struct {
	SessionExpiration int `default:"15"` //in minutes

	LDAPServer       string `required:"true"`
	LDAPPort         int    `default:"389" required:"true"`
	LDAPBaseDN       string `required:"true"`
	LDAPBindUPN      string `required:"true"`
	LDAPBindPassword string `required:"true"`
	LDAPSecurity     string `default:"none" required:"true"`
	ldapSecurity     adauth.SecurityType

	Permissions string `required:"true"`
	permissions map[string][]auth.GradeRange

	SecureTokenKey string `required:"true"`

	ListenAddr string `default:":8080" required:"true"` //addr format used for net.Dial; required
	Prefix     string //url prefix to mount api to without trailing slash
	Debug      bool   `default:"false"` //return debugging information to client
}

var config = &Config{}

func init() {
	err := envconfig.Process("USERBROWSER", config)
	if err != nil {
		log.Fatalln("Error reading configuration from environment:", err)
	}

	switch strings.ToLower(config.LDAPSecurity) {
	case "", "none":
		config.ldapSecurity = adauth.SecurityNone
	case "tls":
		config.ldapSecurity = adauth.SecurityTLS
	case "starttls":
		config.ldapSecurity = adauth.SecurityStartTLS
	default:
		log.Fatalln("Invalid USERBROWSER_LDAPSECURITY:", config.LDAPSecurity)
	}

	permissions, err := parsePermissions(config.Permissions)
	if err != nil {
		log.Fatalln("Invalid USERBROWSER_PERMISSIONS:", err)
	}

	config.permissions = permissions
}
