package e2e

import (
	"github.com/onsi/gomega/gexec"
	"github.com/sky-uk/fluentd-docker/e2e/k8s"
	"github.com/sky-uk/fluentd-docker/e2e/kind"
	"github.com/sky-uk/fluentd-docker/e2e/util"

	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFluentdSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluentd test suite")
}

const (
	kindClusterName = "es-e2e"
	kindConfig      = resourcesPath + "/kind/kind.conf"

	fluentdDockerImage = "skycirrus/fluentd-docker:latest"

	resourcesPath   = "e2e/resources"
	esManifest      = resourcesPath + "/es"
	fluentdManifest = resourcesPath + "/fluentd"
	kibanaManifest  = resourcesPath + "/kibana"
	loggerManifest  = resourcesPath + "/logging-pod.yml"

	elasticsearchPort = 9200
)

var (
	// Only 1 replica expected for each (from kind.conf)
	kindMembers = []string{"control-plane", "worker"}
	kindctl     kind.Kind
	kubectl     k8s.Kubectl
	esForwarder *k8s.PortForwarder
)

var _ = BeforeSuite(func() {
	kindctl = kind.New(kindClusterName, util.Resource(kindConfig), fluentdDockerImage, kindMembers)
	kindctl.CreateCluster()

	kubectl = k8s.New(kindctl.Kubeconfig())

	// Deploy elastic search
	kubectl.ApplyFromPath(util.Resource(esManifest))
	elasticsearchPod := k8s.Pod{
		Kind:      "statefulset",
		Namespace: "kube-system",
		Name:      "elasticsearch",
		Label:     k8s.Label{Name: "k8s-app", Value: "elasticsearch"},
	}
	kubectl.WaitForStatefulSetReady(elasticsearchPod, time.Duration(2*time.Minute))

	kubectl.ApplyFromPath(util.Resource(fluentdManifest))
	kubectl.WaitForDaemonSetReady(k8s.Pod{
		// Deploy custom fluentd
		Namespace: "kube-system",
		Label:     k8s.Label{Name: "k8s-app", Value: "fluentd-es"},
	}, time.Duration(15*time.Minute))

	esForwarder = k8s.NewPortForwarder(elasticsearchPod, elasticsearchPort)
})

var _ = AfterSuite(func() {
	gexec.KillAndWait()
	kubectl.DeleteFromPath(util.Resource(fluentdManifest))
	kubectl.DeleteFromPath(util.Resource(esManifest))
	kindctl.DeleteCluster()
})
