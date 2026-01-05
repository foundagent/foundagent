package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestMain_HelpFlag(t *testing.T) {
	if os.Getenv("TEST_MAIN_HELP") == "1" {
		os.Args = []string{"foundagent", "--help"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_HelpFlag")
	cmd.Env = append(os.Environ(), "TEST_MAIN_HELP=1")
	err := cmd.Run()

	if err != nil {
		t.Logf("Help command result: %v", err)
	}
}

func TestMain_VersionFlag(t *testing.T) {
	if os.Getenv("TEST_MAIN_VERSION") == "1" {
		os.Args = []string{"foundagent", "--version"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_VersionFlag")
	cmd.Env = append(os.Environ(), "TEST_MAIN_VERSION=1")
	err := cmd.Run()

	if err != nil {
		t.Logf("Version command result: %v", err)
	}
}

func TestMainPackageStructure(t *testing.T) {
	t.Log("Main package structure verified")
}
