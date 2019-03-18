package docker

import (
	"fmt"
	"github.com/sky-uk/fluentd-docker/e2e/cli"
	"os/exec"

	. "github.com/onsi/gomega"
)

const fluentdDockerArchive = "fluentd-docker.tar.gz"

func New(image string) Docker {
	return Docker{
		image:   image,
		archive: fluentdDockerArchive,
	}
}

type Docker struct {
	image   string
	archive string
}

func (d *Docker) SaveImage() {
	args := []string{"save",
		d.image,
		cli.ShortFlag("o", d.archive),
	}
	cmd := exec.Command("docker", args...)
	stdout, stderr, err := cli.ExecWithSuccess(cmd, true)
	Expect(err).ToNot(HaveOccurred(), "SaveImage output: %s\nerror:%s", stdout, stderr)

}

func (d *Docker) DeleteImageFile() {
	cmd := exec.Command("rm", d.archive)
	stdout, stderr, err := cli.ExecWithSuccess(cmd, true)
	Expect(err).ToNot(HaveOccurred(), "DeleteImageFile output: %s\nerror:%s", stdout, stderr)
}

func (d *Docker) UploadImageFile(node string) {
	args := []string{"cp",
		d.archive,
		fmt.Sprintf("%s:/%s", node, d.archive),
	}
	cmd := exec.Command("docker", args...)
	stdout, stderr, err := cli.ExecWithSuccess(cmd, true)
	Expect(err).ToNot(HaveOccurred(), "UploadImageFile output: %s\nerror:%s", stdout, stderr)
}

func (d *Docker) LoadImage(node string) {
	args := []string{"exec",
		node,
		"docker",
		"load",
		"-i",
		fmt.Sprintf("/%s", d.archive),
	}
	cmd := exec.Command("docker", args...)
	stdout, stderr, err := cli.ExecWithSuccess(cmd, true)
	Expect(err).ToNot(HaveOccurred(), "LoadImage output: %s\nerror:%s", stdout, stderr)
}

