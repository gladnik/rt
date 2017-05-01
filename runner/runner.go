package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	. "github.com/aerokube/rt/common"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"
	"text/template"
)

var (
	dataDir      string
	rawTemplates string
	rawBuildData string

	runningCommands []int
)

const (
	completedCode = iota
	failedCode
	terminatedCode
	errorCode
)

func init() {
	dataDir = getEnvOrDefault(DataDir, "/")
	rawTemplates = getEnvOrDefault(Templates, "{}")
	rawBuildData = getEnvOrDefault(BuildData, "{}")
}

func getEnvOrDefault(name string, defaultValue string) string {
	env := os.Getenv(name)
	if env == "" {
		return defaultValue
	}
	return env
}

func main() {
	err := chDir(dataDir)
	if err != nil {
		log.Printf("Invalid data directory: %v\n", err)
		os.Exit(errorCode)
	}

	var buildData StandaloneTestCase
	err = json.Unmarshal([]byte(rawBuildData), &buildData)
	if err != nil {
		log.Printf("Failed to parse template data: %v\n", err)
		os.Exit(errorCode)
	}
	testCaseName := buildData.TestCase.Name

	var templates map[string]string
	err = json.Unmarshal([]byte(rawTemplates), &templates)
	if len(templates) > 0 {
		err = generateBuildFiles(templates, buildData)
		if err != nil {
			log.Printf("Can not obtain build file: %v\n", err)
			os.Exit(errorCode)
		}
	}
	execTests(dataDir, testCaseName)
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGCHLD)
	go func() {
		s := (<-sig).(syscall.Signal)
		for _, pid := range runningCommands {
			syscall.Kill(-(pid), s)
		}
		os.Exit(terminatedCode)
	}()
	var status syscall.WaitStatus
	syscall.Wait4(-1, &status, 0, nil)
}

func chDir(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("invalid directory \"%s\": %v\n", dir, err)
	}
	return os.Chdir(dir)
}

func generateBuildFiles(templates map[string]string, buildData StandaloneTestCase) error {
	for tpl, outputFile := range templates {
		t, err := template.ParseFiles(tpl)
		if err != nil {
			return fmt.Errorf("failed to parse template file \"%s\": %v", tpl, err)
		}
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create build file \"%s\": %v", outputFile, err)
		}
		defer f.Close()
		err = t.Execute(f, buildData)
		if err != nil {
			return fmt.Errorf("failed to generate build file \"%s\": %v", outputFile, err)
		}
	}
	return nil
}

func execTests(dataDir string, testCaseName string) {
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	pid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		log.Fatal(err)
	}
	runningCommands = append(runningCommands, pid)
	
	logFile := path.Join(dataDir, fmt.Sprintf("LOG-%s.log", testCaseName))
	f, err := os.Create(logFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	teeWriter := io.MultiWriter(os.Stdout, bufio.NewWriter(f))
	go io.Copy(teeWriter, stdout)
	go io.Copy(teeWriter, stderr)
	
	testsCmdError := cmd.Wait()
	if testsCmdError == nil {
		os.Exit(completedCode)
	} else {
		os.Exit(failedCode)
	}
}