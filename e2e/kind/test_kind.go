package kind

import (
	"fmt"
	"github.com/sky-uk/fluentd-docker/e2e/cli"
	"github.com/sky-uk/fluentd-docker/e2e/docker"
	"github.com/sky-uk/fluentd-docker/e2e/installable"
	"os/exec"
	"path/filepath"
	"strings"

	"net/url"

	. "github.com/onsi/gomega"
)

const (
	kindVersion          = "0.1.0"
)

func New(clusterName, config, fluentdImage string, members []string) Kind {
	k := Kind{
		binary:         "kind",
		clusterName:    clusterName,
		config:         config,
		nodes:          members,
		desiredVersion: kindVersion,
		docker: docker.New(fluentdImage),
	}
	installable.Install(&k)
	return k
}

type Kind struct {
	clusterName    string
	config         string
	desiredVersion string
	nodes          []string
	docker         docker.Docker
	binPath        string
	binary         string
}

func (k *Kind) Nodes() []string {
	var names []string
	for _, node := range k.nodes {
		names = append(names, fmt.Sprintf("kind-%s-%s", k.clusterName, node))
	}
	return names
}

func (k *Kind) CreateCluster(){
	args := []string{"create", "cluster",
		cli.FullFlag("name", k.clusterName),
		cli.FullFlag("config", k.config),
	}

	stdout, stderr, err:= k.runVerbose(args)
	Expect(err).ToNot(HaveOccurred(), "CreateCluster output: %s\nerror:%s", stdout, stderr)
	k.LoadDockerImage()
}

func (k *Kind) DeleteCluster() {
	args := []string{"delete", "cluster",
		cli.FullFlag("name", k.clusterName)}
	stdout, stderr, err:= k.runVerbose(args)
	Expect(err).ToNot(HaveOccurred(), "DeleteCluster output: %s\nerror:%s", stdout, stderr)
}

func (k *Kind) LoadDockerImage() {
	k.docker.SaveImage()
	defer k.docker.DeleteImageFile()
	for _, node := range k.Nodes() {
		k.docker.UploadImageFile(node)
		k.docker.LoadImage(node)
	}
}

func (k *Kind) Kubeconfig() string {
	args := []string{"get", "kubeconfig-path",
		cli.FullFlag("name", k.clusterName),
	}
	stdout, stderr, err:= k.run(args)
	Expect(err).ToNot(HaveOccurred(), "kind Kubeconfig output: %s\nerror:%s", stdout, stderr)
	return strings.TrimSpace(stdout)
}

func (k *Kind) run(args []string) (stdout string, stderr string, err error) {
	return cli.ExecWithSuccess(exec.Command(k.Command(), args...), false)
}

func (k *Kind) runVerbose(args []string) (stdout string, stderr string, err error) {
	return cli.ExecWithSuccess(exec.Command(k.Command(), args...), true)
}

// Installable
func (k *Kind) ReleaseURL() *url.URL {
	kindURL, err := url.Parse("https://github.com/kubernetes-sigs/kind/releases/download/" + k.desiredVersion + "/kind-linux-amd64")
	Expect(err).ToNot(HaveOccurred(), "Kind ReleaseURL")
	return kindURL
}

func (k *Kind) BinaryName() string {
	return k.binary
}

func (k *Kind) BinaryPath() string {
	return k.binPath
}

func (k *Kind) SetBinaryPath(path string) {
	k.binPath = path
}

func (k *Kind) Command() string {
	return filepath.Join(k.BinaryPath(), k.BinaryName())
}

func (k *Kind) CurrentVersion() string {
	args := []string{"version"}
	stdout, stderr, err:= k.run(args)
	Expect(err).ToNot(HaveOccurred(), "kind CurrentVersion output: %s\nerror:%s", stdout, stderr)
	return stderr
}

func (k *Kind) DesiredVersion() string {
	return k.desiredVersion
}
