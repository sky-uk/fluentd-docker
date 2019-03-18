package e2e

import (
	"context"
	"gopkg.in/olivere/elastic.v6"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("elasticsearch", func() {

	var client *elastic.Client

	JustBeforeEach(func() {
		var err error
		kubectl.ForwardPort(esForwarder)
		client, err = elastic.NewClient(
			elastic.SetURL(esForwarder.URL().String()),
			elastic.SetSniff(false),
		)
		Expect(err).ToNot(HaveOccurred(), "should create elasticsearch client for %s", esForwarder.URL())
	})

	AfterEach(func() {
		client.Stop()
	})

	It("is healthy", func() {
		resp, err := client.CatHealth().Columns("status").Do(context.Background())
		Expect(err).ToNot(HaveOccurred(), "should query cluster health for %s", esForwarder.URL())
		Expect(resp).ToNot(BeEmpty(), "The cluster health should have at least one entry for %s", esForwarder.URL())
		for _, row := range resp {
			Expect(row.Status).To(Equal("green"))
		}
	})
})
