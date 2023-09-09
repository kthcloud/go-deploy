package harbor

import "testing"

func TestCreateRobot(t *testing.T) {
	project := withHarborProject(t)
	robot := withHarborRobot(t, project)
	cleanUpHarborRobot(t, robot.ID)
	cleanUpHarborProject(t, project.ID)
}
