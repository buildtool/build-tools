package version

import (
	"fmt"
)

type Info struct {
	Name        string
	Description string
	Version     string
	Commit      string
	Date        string
}

func (v Info) String() string {
	return fmt.Sprintf("Version: %v, commit %v, built at %v\n", v.Version, v.Commit, v.Date)
}
