package installable

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"strings"

	"io/ioutil"
	"net/url"
	"os/exec"
	"path/filepath"

	"github.com/sky-uk/fluentd-docker/e2e/cli"
)

type Installable interface {
	BinaryName() string
	SetBinaryPath(path string)
	BinaryPath() string
	CurrentVersion() string
	DesiredVersion() string
	ReleaseURL() *url.URL
}

func Installed(item Installable) (string, error) {
	cmd := exec.Command("which", item.BinaryName())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return filepath.Dir(strings.TrimSpace(string(out))), nil
}

func Install(item Installable) {
	path, err := Installed(item)
	item.SetBinaryPath(path)
	if err != nil || !strings.Contains(item.CurrentVersion(), item.DesiredVersion()) {
		path, err = ioutil.TempDir("", fmt.Sprintf("%s_%s", item.BinaryName(), item.DesiredVersion()))
		item.SetBinaryPath(path)
		_,_= ginkgo.GinkgoWriter.Write([]byte(fmt.Sprintf("\nInstalling %s::%s at %s\n", item.BinaryName(), item.DesiredVersion(), item.BinaryPath())))
		stdout, stderr, err := DownloadAndInstall(item)
		Expect(err).ToNot(HaveOccurred(), "Install output: %s\nerror:%s", stdout, stderr)
	}
}

func DownloadAndInstall(k Installable) (stdout string, stderr string, err error) {
	cmd := exec.Command("curl", "-Lo", k.BinaryName(), k.ReleaseURL().String())
	cmd.Dir = k.BinaryPath()
	stdout, stderr, err = cli.ExecWithSuccess(cmd, true)
	if err!= nil {
		return stdout, stderr, err
	}
	cmd = exec.Command("chmod", "+x", k.BinaryName())
	cmd.Dir = k.BinaryPath()
	stdout, stderr, err = cli.ExecWithSuccess(cmd, true)
	if err!= nil {
		return stdout, stderr, err
	}
	return "","",nil
}
