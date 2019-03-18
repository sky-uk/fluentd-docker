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
)

var (
	// Only 1 replica expected for each (from kind.conf)
	kindMembers = []string{"control-plane", "worker"}
	elasticsearch k8s.PortForwarder
	kindctl     kind.Kind
	kubectl     k8s.Kubectl
)

var _ = BeforeSuite(func() {
	kindctl = kind.New(kindClusterName, util.Resource(kindConfig), fluentdDockerImage, kindMembers)
	kindctl.CreateCluster()

	kubectl = k8s.New(kindctl.Kubeconfig())

	// Deploy elastic search
	kubectl.ApplyFromPath(util.Resource(esManifest))
	elasticsearchPod := k8s.Pod{
		Namespace:     "kube-system",
		Label:         k8s.Label{Name: "k8s-app", Value: "elasticsearch"},
		PortForwarder: k8s.NewStatefulSetPortForwarder("elasticsearch", 9200),
	}
	kubectl.WaitForStatefulSetReady(elasticsearchPod, time.Duration(2*time.Minute))

	kubectl.ApplyFromPath(util.Resource(fluentdManifest))
	kubectl.WaitForDaemonSetReady(k8s.Pod{
		// Deploy custom fluentd
		Namespace: "kube-system",
		Label:     k8s.Label{Name: "k8s-app", Value: "fluentd-es"},
	}, time.Duration(1*time.Minute))

	//ES takes a bit before being ready. TODO: try set up readiness instead
	time.Sleep(15*time.Second)
	kubectl.ForwardPort(elasticsearchPod)
	elasticsearch = elasticsearchPod.PortForwarder
})

var _ = AfterSuite(func() {
	gexec.KillAndWait()
	kubectl.DeleteFromPath(util.Resource(fluentdManifest))
	kubectl.DeleteFromPath(util.Resource(esManifest))
	kindctl.DeleteCluster()
})
