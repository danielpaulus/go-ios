package main_test

import (
	"flag"
	"fmt"
	"os/exec"
	"testing"
)

var (
	update = flag.Bool("update", false, "update golden files")
	e2e    = flag.Bool("e2e", false, "test with realdevice")
)

func TestDeviceList(t *testing.T) {
	if !*e2e {
		return
	}
	output, err := exec.Command("go", "run", "./ios.go", "list").Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(output))
}
