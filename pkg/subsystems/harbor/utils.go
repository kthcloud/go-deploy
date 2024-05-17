package harbor

import (
	"fmt"
)

// getRobotFullName returns the full name of a robot.
// This is used because Harbor returns a prepended name (default: robot$) when listing robots.
func getRobotFullName(projectName, name string) string {
	return fmt.Sprintf("robot$%s", getRobotName(projectName, name))
}

// getRobotName returns the name of a robot.
// This is used because Harbor returns a prepended name (default: robot$) when listing robots.
func getRobotName(projectName, name string) string {
	return fmt.Sprintf("%s+%s", projectName, name)
}
