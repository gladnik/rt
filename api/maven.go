package api

import (
	"fmt"
	. "github.com/aerokube/rt/common"
	"github.com/aerokube/rt/config"
	"path"
)

type MavenTool struct {
}

func (mt *MavenTool) GetCommand(container *config.Container, testCase TestCase, properties []Property) Command {
	//mvn -f pom.xml [...properties] verify
	pomXmlPath := path.Join(container.DataDir, "pom.xml")
	cmd := []string{"mvn", "-f", pomXmlPath, "-Dmaven.repo.local=/root/.m2", "-Dtest=" + testCase.Name}
	for _, property := range properties {
		cmd = append(cmd, fmt.Sprintf("-D%s=%s", property.Key, property.Value))
	}
	cmd = append(cmd, "verify")
	return cmd
}
