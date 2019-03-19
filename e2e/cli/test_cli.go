package cli

import (
	"bytes"
	"fmt"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gexec"
	"io"
	"os/exec"
	"strings"
)

func FullFlag(name, value string) string {
	return Flag("--", name, value)
}

func ShortFlag(name, value string) string {
	return Flag("-", name, value)
}

func Flag(prefix, name, value string) string {
	return fmt.Sprintf("%s%s=%s", prefix, name, value)
}

func ExecWithSuccess(cmd *exec.Cmd, verbose bool) (stdout string, stderr string, err error) {
	var stderrB bytes.Buffer
	var stdoutB bytes.Buffer
	errWriters := []io.Writer{&stderrB}
	outWriters := []io.Writer{&stdoutB}
	if verbose {
		errWriters = append(errWriters, ginkgo.GinkgoWriter)
		outWriters = append(outWriters, ginkgo.GinkgoWriter)
	}
	cmd.Stderr = io.MultiWriter(errWriters...)
	cmd.Stdout = io.MultiWriter(outWriters...)
	err = cmd.Run()
	stdout = stdoutB.String()
	stderr = stderrB.String()
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(stdout), strings.TrimSpace(stderr), nil
}

func ExecAsync(cmd *exec.Cmd) (*gexec.Session, error) {
	session, err := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	return session, err
}