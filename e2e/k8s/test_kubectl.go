package k8s

import (
	"fmt"
	"github.com/onsi/gomega/gexec"
	"github.com/sky-uk/fluentd-docker/e2e/cli"
	"github.com/sky-uk/fluentd-docker/e2e/installable"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/gomega"
)

const (
	kubectlVersion = "v1.10.0"
)

func New(kubeconfig string) Kubectl {
	k := Kubectl{
		kubeconfig:     kubeconfig,
		binary:         "kubectl",
		desiredVersion: kubectlVersion,
	}
	installable.Install(&k)
	k.createClient()
	return k
}

type Kubectl struct {
	kubeconfig     string
	binary         string
	binPath        string
	desiredVersion string
	client         *kubernetes.Clientset
}

func (k *Kubectl) ApplyFromPath(path string) {
	args := []string{"apply",
		cli.ShortFlag("f", path),
	}
	stdout, stderr, err := k.runVerbose(args)
	Expect(err).ToNot(HaveOccurred(), "ApplyFromPath %s\noutput: %s\nerror:%s", path, stdout, stderr)
}

func (k *Kubectl) DeleteFromPath(path string) {
	args := []string{"delete",
		cli.FullFlag("kubeconfig", k.kubeconfig),
		cli.ShortFlag("f", path),
	}
	stdout, stderr, err := k.runVerbose(args)
	Expect(err).ToNot(HaveOccurred(), "DeleteFromPath %s\noutput: %s\nerror:%s", path, stdout, stderr)
}

func (k *Kubectl) createClient() {
	config, err := clientcmd.BuildConfigFromFlags("", k.kubeconfig)
	Expect(err).ToNot(HaveOccurred(), "should load the KinD kubeconfig:s", k.kubeconfig)
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	Expect(err).ToNot(HaveOccurred(), "should create the KinD client")
	k.client = clientset
}

func (k *Kubectl) run(args []string) (stdout string, stderr string, err error) {
	withKubeconfig := k.prependKubeconfig(args)
	return cli.ExecWithSuccess(exec.Command(k.Command(), withKubeconfig...), false)
}

func (k *Kubectl) runAsync(args []string) (*gexec.Session, error) {
	withKubeconfig := k.prependKubeconfig(args)
	return cli.ExecAsync(exec.Command(k.Command(), withKubeconfig...))
}

func (k *Kubectl) runVerbose(args []string) (stdout string, stderr string, err error) {
	withKubeconfig := k.prependKubeconfig(args)
	return cli.ExecWithSuccess(exec.Command(k.Command(), withKubeconfig...), true)
}

func (k *Kubectl) prependKubeconfig(args []string) []string {
	withKubeconfig := []string{cli.FullFlag("kubeconfig", k.kubeconfig)}
	return append(withKubeconfig, args...)
}

type Pod struct {
	Namespace     string
	PortForwarder PortForwarder
	Label         Label
}

type Label struct {
	Name  string
	Value string
}

func (l *Label) Selector() string {
	return fmt.Sprintf("%s=%s", l.Name, l.Value)
}

func (k *Kubectl) WaitForStatefulSetReady(pod Pod, timeout time.Duration) {
	err := withRetry(timeout, func() (completed bool) {
		selector := metav1.ListOptions{LabelSelector: pod.Label.Selector()}
		statefulSets, err := k.client.AppsV1().StatefulSets(pod.Namespace).List(selector)
		if err != nil {
			if unrecoverable(err) == nil {
				return false
			}
			Expect(err).ToNot(HaveOccurred(), "Unexpected error", pod.Label.Selector())
		}
		for _, statefulSet := range statefulSets.Items {
			if *statefulSet.Spec.Replicas == statefulSet.Status.ReadyReplicas &&
				statefulSet.Status.ObservedGeneration == statefulSet.Generation &&
				statefulSet.Status.CurrentRevision == statefulSet.Status.UpdateRevision {
				fmt.Printf("StatefulSet %q is ready.\n", pod.Label.Selector())
				return true
			}
		}
		return false
	})
	Expect(err).ToNot(HaveOccurred(), "StatefulSet %q is not ready", pod.Label.Selector())
}

func (k *Kubectl) WaitForDaemonSetReady(pod Pod, timeout time.Duration) {
	err := withRetry(timeout, func() (completed bool) {
		selector := metav1.ListOptions{LabelSelector: pod.Label.Selector()}
		daemonSets, err := k.client.AppsV1().DaemonSets(pod.Namespace).List(selector)
		if err != nil {
			if unrecoverable(err) == nil {
				return false
			}
			Expect(err).ToNot(HaveOccurred(), "Unexpected error", pod.Label.Selector())
		}
		for _, daemonSet := range daemonSets.Items {
			if daemonSet.Generation == daemonSet.Status.ObservedGeneration &&
				daemonSet.Status.NumberUnavailable == 0 &&
				daemonSet.Status.UpdatedNumberScheduled > 0 &&
				daemonSet.Status.UpdatedNumberScheduled == daemonSet.Status.DesiredNumberScheduled {
				fmt.Printf("DaemonSet %q is ready.\n", pod.Label.Selector())
				return true
			}
		}
		return false
	})
	Expect(err).ToNot(HaveOccurred(), "StatefulSet %q is not ready", pod.Label.Selector())
}

func withRetry(timeout time.Duration, task func() bool) error {
	var elapsedTime time.Duration
	const waitTime = 10 * time.Second
	for {
		if task() {
			return nil
		}
		if elapsedTime >= timeout {
			return fmt.Errorf("timed out after %s", elapsedTime)
		}
		time.Sleep(waitTime)
		elapsedTime += waitTime
	}
}

func unrecoverable(err error) error {
	if errors.IsNotFound(err) {
		fmt.Printf(".")
		return nil
	}
	return err
}

type PortForwarder interface {
	Kind() string
	Name() string
	SetLocalPort(port int32)
	LocalPort() int32
	ContainerPort() int32
	SetSession(session *gexec.Session)
	Stop()
	IsForwarding() bool
	URL() *url.URL
}

func NewStatefulSetPortForwarder(name string, port int32) PortForwarder {
	return &statefulsetPortForwarder{
		name:          name,
		containerPort: port,
	}
}

type statefulsetPortForwarder struct {
	name          string
	containerPort int32
	localPort     int32
	session       *gexec.Session
}

func (spf *statefulsetPortForwarder) Kind() string {
	return "statefulset"
}

func (spf *statefulsetPortForwarder) Name() string {
	return spf.name
}

func (spf *statefulsetPortForwarder) URL() *url.URL {
	location, err:= url.Parse(fmt.Sprintf("http://localhost:%d", spf.LocalPort()))
	Expect(err).ToNot(HaveOccurred(), "PortForwarder URL")
	return location
}

func (spf *statefulsetPortForwarder) ContainerPort() int32 {
	return spf.containerPort
}

func (spf *statefulsetPortForwarder) LocalPort() int32 {
	return spf.localPort
}

func (spf *statefulsetPortForwarder) SetLocalPort(port int32) {
	spf.localPort = port
}

func (spf *statefulsetPortForwarder) SetSession(session *gexec.Session) {
	spf.session = session
}

func (spf *statefulsetPortForwarder) Stop() {
	spf.session.Terminate().Wait()
}

func (spf *statefulsetPortForwarder) IsForwarding() bool {
	return spf.session.ExitCode() == -1
}

func (k *Kubectl) ForwardPort(pod Pod) {
	args := []string{"port-forward",
		cli.FullFlag("namespace", pod.Namespace),
		fmt.Sprintf("%s/%s", pod.PortForwarder.Kind(), pod.PortForwarder.Name()),
		fmt.Sprintf(":%d", pod.PortForwarder.ContainerPort()),
	}
	session, err := k.runAsync(args)
	Expect(err).ToNot(HaveOccurred(), "Request ForwardPort for %s.%s:%d", pod.Namespace, pod.PortForwarder.Name(), pod.PortForwarder.ContainerPort())
	pod.PortForwarder.SetSession(session)
	fpRegexp := regexp.MustCompile(`(?m)[^:]+:(\d+) .*$`)
	err = withRetry(30*time.Second, func() bool {
		return strings.Contains(string(session.Out.Contents()), "Forwarding from")
	})
	Expect(err).ToNot(HaveOccurred(), "Start ForwardPort for %s.%s:%d", pod.Namespace, pod.PortForwarder.Name(), pod.PortForwarder.ContainerPort())
	matches := fpRegexp.FindStringSubmatch(string(session.Out.Contents()))
	Expect(matches).To(HaveLen(2), "Should match forwarding port regexp")
	localPort, err := strconv.ParseInt( matches[1], 0, 32)
	Expect(err).ToNot(HaveOccurred(), "Local port should be an integer")
	pod.PortForwarder.SetLocalPort(int32(localPort))
}

// Installable
func (k *Kubectl) ReleaseURL() *url.URL {
	kubectlURL, err := url.Parse("https://storage.googleapis.com/kubernetes-release/release/" + k.desiredVersion + "/bin/linux/amd64/kubectl")
	Expect(err).ToNot(HaveOccurred(), "k8s ReleaseURL")
	return kubectlURL
}

func (k *Kubectl) BinaryName() string {
	return k.binary
}

func (k *Kubectl) BinaryPath() string {
	return k.binPath
}

func (k *Kubectl) SetBinaryPath(path string) {
	k.binPath = path
}

func (k *Kubectl) Command() string {
	return filepath.Join(k.BinaryPath(), k.BinaryName())
}

func (k *Kubectl) CurrentVersion() string {
	args := []string{"version", "--client", "--short"}
	stdout, stderr, err := k.run(args)
	Expect(err).ToNot(HaveOccurred(), "k8s Kubeconfig output: %s\nerror:%s", stdout, stderr)
	return stdout
}

func (k *Kubectl) DesiredVersion() string {
	return k.desiredVersion
}
