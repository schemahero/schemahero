package types

import (
	"github.com/schemahero/schemahero/pkg/client/schemaheroclientset/scheme"
)

type Spec struct {
	SourceFilename string
	Spec           []byte
}

type Specs []Spec

// Spec sort interface implementation
func (s Specs) Len() int {
	return len(s)
}

func (s Specs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Ensure tables are processed before views, secondary sort is by file name
func (s Specs) Less(i, j int) bool {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	_, gvkI, err := decode(s[i].Spec, nil, nil)
	if err != nil {
		return s[i].SourceFilename < s[j].SourceFilename
	}

	_, gvkJ, err := decode(s[j].Spec, nil, nil)
	if err != nil {
		return s[i].SourceFilename < s[j].SourceFilename
	}

	if gvkI.Group == "schemas.schemahero.io" && gvkJ.Group == "schemas.schemahero.io" {
		if gvkI.Kind == gvkJ.Kind {
			return s[i].SourceFilename < s[j].SourceFilename
		}
		if gvkI.Kind == "Table" {
			return true
		}
		if gvkJ.Kind == "Table" {
			return false
		}
	}

	return s[i].SourceFilename < s[j].SourceFilename
}
