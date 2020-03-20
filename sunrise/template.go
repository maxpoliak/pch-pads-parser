package sunrise

import (
	"strings"
)

// KeywordCheck - This function is used to filter parsed lines of the configuration file and
//                returns true if the keyword is contained in the line.
// line      : string from the configuration file
func KeywordCheck(line string) bool {
	for _, keyword := range []string{
		"GPP_", "GPD",
	} {
		if strings.Contains(line, keyword) {
			return true
		}
	}
	return false
}