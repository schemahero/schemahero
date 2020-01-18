package v1alpha3

type ValueOrValueFrom struct {
	Value     string     `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFrom *ValueFrom `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}
