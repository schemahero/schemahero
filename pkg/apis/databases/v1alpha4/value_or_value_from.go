package v1alpha4

import (
	"github.com/pkg/errors"
)

type ValueOrValueFrom struct {
	Value     string     `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFrom *ValueFrom `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}

// IsEmpty returns true if there is not a value in value and valuefrom
func (v *ValueOrValueFrom) IsEmpty() bool {
	if v.Value != "" {
		return false
	}

	if v.ValueFrom != nil {
		return false
	}

	return true
}

// HasVaultSecret returns true if the ValueOrValueFrom
// contains a Vault stanza
func (v *ValueOrValueFrom) HasVaultSecret() bool {
	if v.ValueFrom != nil {
		return v.ValueFrom.Vault != nil
	}
	return false
}

// GetVaultDetails returns the configured Vault details for the
// ValueOrValueFrom, or returns error if Vault stanza is missing
func (v *ValueOrValueFrom) GetVaultDetails() (*Vault, error) {
	if v.HasVaultSecret() {
		return v.ValueFrom.Vault, nil
	}

	return nil, errors.New("No Vault secret configured")
}
