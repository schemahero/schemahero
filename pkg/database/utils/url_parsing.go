package utils

import (
	"net"
	nurl "net/url"

	"github.com/pkg/errors"
)

// UsernameFromURL extracts the username from a URL string.
// Returns an empty string if no username is present.
func UsernameFromURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	if u.User == nil {
		return "", nil
	}
	return u.User.Username(), nil
}

// PasswordFromURL extracts the password from a URL string.
// Returns an empty string if no password is present.
func PasswordFromURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	if u.User == nil {
		return "", nil
	}
	pass, _ := u.User.Password()
	return pass, nil
}

// HostnameFromURL extracts the hostname from a URL string.
func HostnameFromURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", errors.Wrap(err, "failed to split host port")
	}
	return host, nil
}

// PortFromURL extracts the port from a URL string.
func PortFromURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", errors.Wrap(err, "failed to split host port")
	}
	return port, nil
}
