package api

import "fmt"

type MavenTool struct {
}

func (mt *MavenTool) GetSettings(testCase TestCase, properties []Property) ToolSettings {
	command := []string{"mvn", "-f", "pom.xml", "-Dtest=" + testCase.Name}
	for _, property := range properties {
		command = append(command, fmt.Sprintf("-D%s=%s", property.Key, property.Value))
	}
	return ToolSettings{
		Command:   command,
		BuildData: make(map[string]string),
	}
}
