package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os/exec"
)

var _ = Describe("elasticsearch", func() {

	It("is running", func() {
		location := elasticsearch.URL().String() + "/_cluster/health"
		cmd:= exec.Command("curl", location)
		out , err:=cmd.CombinedOutput()
		Expect(err).ToNot(HaveOccurred(),"Can curl %#v: \n%s\nerror: %v", cmd, out, err)
		Expect(out).To(ContainSubstring("\"status\":\"green\""))
	})
})
