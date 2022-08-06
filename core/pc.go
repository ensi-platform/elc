package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"

	"github.com/mattn/go-isatty"
)

type PC interface {
	ExecInteractive(command []string, env []string) (int, error)
	ExecToString(command []string, env []string) (int, string, error)
	Args() []string
	Exit(code int)
	HomeDir() (string, error)
	Getuid() int
	Getwd() (dir string, err error)
	FileExists(filepath string) bool
	ReadFile(filename string) ([]byte, error)
	ReadDir(dirname string) ([]os.FileInfo, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	Printf(format string, a ...interface{}) (n int, err error)
	Println(a ...interface{}) (n int, err error)
	IsTerminal() bool
}

var Pc PC

type RealPC struct{}

func (r *RealPC) ExecInteractive(command []string, env []string) (int, error) {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), env...)

	err := cmd.Run()

	return cmd.ProcessState.ExitCode(), err
}

func (r *RealPC) ExecToString(command []string, env []string) (int, string, error) {
	var stdout bytes.Buffer
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), env...)

	err := cmd.Run()
	return cmd.ProcessState.ExitCode(), stdout.String(), err
}

func (r *RealPC) Args() []string {
	return os.Args
}

func (r *RealPC) Exit(code int) {
	os.Exit(code)
}

func (r *RealPC) HomeDir() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", nil
	}

	return currentUser.HomeDir, nil
}

func (r *RealPC) Getuid() int {
	return os.Getuid()
}

func (r *RealPC) Getwd() (dir string, err error) {
	return os.Getwd()
}

func (r *RealPC) FileExists(filepath string) bool {
	_, err := os.Stat(filepath)

	return err == nil
}

func (r *RealPC) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func (r *RealPC) ReadDir(dirname string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

func (r *RealPC) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func (r *RealPC) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Printf(format, a...)
}

func (r *RealPC) Println(a ...interface{}) (n int, err error) {
	return fmt.Println(a...)
}

func (r *RealPC) IsTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}
