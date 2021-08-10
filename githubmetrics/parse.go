package githubmetrics

import (
	"errors"
	"strings"
)

// ParseRepoID expects a string with format "owner/repo" and will return
// owner and repo separately. Returns error on unexpected format.
func ParseRepoID(str string) (owner string, repo string, err error) {
	parts := strings.Split(str, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New(`incorrect repository format. expected "<owner>/<repo>"`)
	}
	return parts[0], parts[1], nil

}
