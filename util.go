package pgsql

import (
	"http"
	"os"
	"strings"
)

import "C" // Redundant in this file, but cgo complains otherwise.

// Parses a URL and returns a Postgres connection string.
// Format: postgres://<user>:<pass>@<host>:<port>/<dbname>
func ParseURL(url string) (cs string, err os.Error) {
	var u *http.URL

	u, err = http.ParseURL(url)
	if err != nil {
		return "", err
	}

	// Add dbname if present
	if len(u.Path) > 1 {
		cs += "dbname=" + u.Path[1:]
	} else {
		return "", ErrMissingDbName
	}

	// Add user if present
	up := strings.Split(u.RawUserinfo, ":")
	if len(up) > 0 && up[0] != "" {
		cs += " user=" + up[0]
	}

	// Add password if present
	if len(up) > 1 && up[1] != "" {
		cs += " password=" + up[1]
	}

	// Add host if present
	hp := strings.Split(u.Host, ":")
	if len(hp) > 0 {
		cs += " host=" + hp[0]
	}

	// Add port if present
	if len(hp) > 1 {
		cs += " port=" + hp[1]
	}

	return cs, nil
}
