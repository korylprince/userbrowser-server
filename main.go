package main

import (
	"log"
	"net/http"
	"os"
	"time"

	adauth "github.com/korylprince/go-ad-auth/v3"
	"github.com/korylprince/userbrowser-server/v3/auth/ad"
	"github.com/korylprince/userbrowser-server/v3/db/ldap"
	"github.com/korylprince/userbrowser-server/v3/httpapi"
	"github.com/korylprince/userbrowser-server/v3/session/memory"
)

func main() {
	authConfig := &adauth.Config{
		Server:   config.LDAPServer,
		Port:     config.LDAPPort,
		BaseDN:   config.LDAPBaseDN,
		Security: config.ldapSecurity,
	}

	db := ldap.New(authConfig, config.LDAPBindUPN, config.LDAPBindPassword, config.SecureTokenKey, config.Debug)
	auth := ad.New(authConfig, config.permissions)
	sessionStore := memory.New(time.Minute * time.Duration(config.SessionExpiration))

	httpapi.Debug = config.Debug
	s := httpapi.NewServer(db, auth, sessionStore, os.Stdout)

	log.Println("Listening on:", config.ListenAddr)

	log.Println(http.ListenAndServe(config.ListenAddr, http.StripPrefix(config.Prefix, s.Router())))
}
