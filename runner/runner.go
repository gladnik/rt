package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"flag"
	"fmt"
	"path"
)

var (
	dataDir string
	testCaseName string
	runningCommands []int
)

const (
	completedCode = iota
	failedCode
	terminatedCode
	errorCode
)

func init() {
	flag.StringVar(&dataDir, "data-dir", "/data", "directory to output data to")
	flag.StringVar(&testCaseName, "test-case", "", "test case name")
	flag.Parse()
}

func main() {
	err := chDir(dataDir)
	if (err != nil) {
		log.Printf("Invalid data directory: %v\n", err)
		os.Exit(errorCode)
	} 
	if (testCaseName == "") {
		log.Println("Test case name can not be empty")
		os.Exit(errorCode)
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
		return fmt.Errorf("Invalid directory \"%s\": %v\n", dir, err)
	}
	return os.Chdir(dir)
}

func execTests(dataDir string, testCaseName string) {
	testsCmd, _ := execCommand(os.Args[1], os.Args[2:]...)
	stdout, err := testsCmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := testsCmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	logFile := path.Join(dataDir, fmt.Sprintf("LOG-%s.log", testCaseName))
	f, err := os.Create(logFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	teeWriter := io.MultiWriter(os.Stdout, bufio.NewWriter(f))
	go io.Copy(teeWriter, stdout)
	go io.Copy(teeWriter, stderr)
	go func() {
		testsCmdError := testsCmd.Wait()
		if testsCmdError == nil {
			os.Exit(completedCode)
		} else {
			os.Exit(failedCode)
		}
	}()
}

func execCommand(name string, arg ...string) (*exec.Cmd, int) {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	pid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		log.Fatal(err)
	}
	runningCommands = append(runningCommands, pid)
	return cmd, pid
}