package v1alpha1

import (
	"testing"

	"github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

func TestUnmarshalInlineOrRef(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type testStruct struct {
		Test InlineOrRef `yaml:"test"`
	}

	inlineYAML := `test: "test"`
	var inline testStruct
	err := yaml.Unmarshal([]byte(inlineYAML), &inline)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(inline.Test.Value).To(gomega.Equal("test"))
	g.Expect(inline.Test.ValueFrom).To(gomega.BeNil())

	secretYAML := `test:
  valueFrom:
    secretKeyRef:
      name: secretName
      key: secretKey`
	var secret testStruct
	err = yaml.Unmarshal([]byte(secretYAML), &secret)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(secret.Test.Value).To(gomega.Equal(""))
	g.Expect(secret.Test.ValueFrom).ToNot(gomega.BeNil())
	g.Expect(secret.Test.ValueFrom.SecretKeyRef).ToNot(gomega.BeNil())
	g.Expect(secret.Test.ValueFrom.SecretKeyRef.Name).To(gomega.Equal("secretName"))
	g.Expect(secret.Test.ValueFrom.SecretKeyRef.Key).To(gomega.Equal("secretKey"))
}
