package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if !IsDebianUbuntuOrAlpine() {
		fail("only ubuntu, debian or alpine are supported")
	}
	if !DpkgExists() {
		fail("dpkg needs to be installed to check dependencies")
	}

	checkDep("libusb-1.0-0-dev")
	checkDep("build-essential")
	checkDep("pkg-config")

	log.Println("good to go. run 'make'")
	//apt-get install -y libusb-1.0-0-dev

}

func checkDep(dep string) {
	log.Println("checking: " + dep)
	if !CheckPackageInstalled(dep) {
		log.Println("installing: " + dep)
		InstallPackage(dep)
	}
	log.Println("ok: " + dep)
}

func fail(reason string) {
	panic(reason)
}

// CheckPackageInstalled checks if the specified package is installed.
// It uses dpkg on Debian/Ubuntu and apk on Alpine.
func CheckPackageInstalled(packageName string) bool {
	if DpkgExists() {
		output := ExecuteCommand("dpkg", "-l", packageName)
		return strings.Contains(output, packageName) && !strings.Contains(output, "no packages found")
	} else if ApkExists() {
		output := ExecuteCommand("apk", "info", packageName)
		return output != ""
	}
	log.Println("No compatible package manager found (dpkg or apk).")
	return false
}

// ExecuteCommand runs the specified shell command and returns its output or panics if there's an error.
func ExecuteCommand(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("Failed to execute command: %s\nError: %v, Stderr: %v\n", command, err, stderr.String())
	}
	return out.String()
}

// DpkgExists checks if dpkg exists in the system's PATH.
func DpkgExists() bool {
	_, err := exec.LookPath("dpkg")
	return err == nil
}

// IsDebianUbuntuOrAlpine checks if the operating system is Debian, Ubuntu, or Alpine.
func IsDebianUbuntuOrAlpine() bool {
	content, err := os.ReadFile("/etc/os-release")
	if err != nil {
		log.Printf("Failed to read /etc/os-release: %v", err)
		return false
	}

	osRelease := string(content)
	return strings.Contains(osRelease, "ID=debian") || strings.Contains(osRelease, "ID=ubuntu") || strings.Contains(osRelease, "ID=alpine")
}

func IsRunningWithSudo() bool {
	// The effective user ID (eUID) is 0 for the root user.
	return os.Geteuid() == 0
}

// PackageManagerType returns the type of package manager available on the system.
func PackageManagerType() string {
	if _, err := exec.LookPath("apt-get"); err == nil {
		return "apt-get"
	} else if _, err := exec.LookPath("apk"); err == nil {
		return "apk"
	}
	return ""
}

// InstallPackage installs a package using the system's package manager.
func InstallPackage(packageName string) {
	var command string
	var args []string

	// Check if apt-get is available
	if _, err := exec.LookPath("apt-get"); err == nil {
		command = "apt-get"
		args = []string{"install", "-y", packageName}
	} else if _, err := exec.LookPath("apk"); err == nil {
		// If apt-get is not available, check for apk
		command = "apk"
		args = []string{"add", packageName}
	} else {
		log.Panic("No compatible package manager found (apt-get or apk).")
	}

	// Execute the install command
	log.Printf("Installing package %s using %s\n", packageName, command)
	ExecuteCommand(command, args...)
}

// ApkExists checks if apk (Alpine Package Keeper) exists in the system's PATH.
func ApkExists() bool {
	_, err := exec.LookPath("apk")
	return err == nil
}
