package api

import (
	"fmt"
	. "github.com/aerokube/rt/common"
)

type MavenTool struct {
}

func (mt *MavenTool) GetSettings(testCase TestCase, properties []Property) Command {
	//mvn -f pom.xml [...properties] verify
	cmd := []string{"mvn", "-f", "pom.xml", "-Dtest=" + testCase.Name}
	for _, property := range properties {
		cmd = append(cmd, fmt.Sprintf("-D%s=%s", property.Key, property.Value))
	}
	cmd = append(cmd, "verify")
	return cmd
}
