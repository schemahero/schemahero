package v1alpha3

import (
	"github.com/schemahero/schemahero/pkg/apis/databases/v1alpha2"
)

type ValueOrValueFrom struct {
	Value     string     `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFrom *ValueFrom `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}

func ConvertValueOrValueFromFromV1Alpha2(instance v1alpha2.ValueOrValueFrom) ValueOrValueFrom {
	converted := ValueOrValueFrom{
		Value: instance.Value,
	}

	if instance.ValueFrom != nil {
		valueFrom := ValueFrom{}
		if instance.ValueFrom.SecretKeyRef != nil {
			valueFrom.SecretKeyRef = &SecretKeyRef{
				Name: instance.ValueFrom.SecretKeyRef.Name,
				Key:  instance.ValueFrom.SecretKeyRef.Key,
			}
		}

		converted.ValueFrom = &valueFrom
	}

	return converted
}
