package v1alpha4

type ValueOrSecretRef struct {
	Value     string           `json:"value" yaml:"value"`
	ValueFrom *ValueFromSecret `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}

type ValueFrom struct {
	SecretKeyRef *SecretKeyRef `json:"secretKeyRef,omitempty" yaml:"secretKeyRef,omitempty"`
	Vault        *Vault        `json:"vault,omitempty" yaml:"vault,omitempty"`
	SSM          *SSM          `json:"ssm,omitempty" yaml:"ssm,omitempty"`
}

type ValueFromSecret struct {
	SecretKeyRef *SecretKeyRef `json:"secretKeyRef,omitempty" yaml:"secretKeyRef,omitempty"`
}

type SecretKeyRef struct {
	Name string `json:"name" yaml:"name"`
	Key  string `json:"key" yaml:"key"`
}

type Vault struct {
	AgentInject             bool   `json:"agentInject,omitempty" yaml:"agentInject,omitempty"`
	Secret                  string `json:"secret" yaml:"secret"`
	Role                    string `json:"role" yaml:"role"`
	Endpoint                string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	ServiceAccount          string `json:"serviceAccount,omitempty" yaml:"serviceAccount,omitempty"`
	ServiceAccountNamespace string `json:"serviceAccountNamespace,omitempty" yaml:"serviceAccountNamespace,omitempty"`
	ConnectionTemplate      string `json:"connectionTemplate,omitempty" yaml:"connectionTemplate,omitempty"`
	KubernetesAuthEndpoint  string `json:"kubernetesAuthEndpoint,omitempty" yaml:"kubernetesAuthEndpoint,omitempty"`
}

type SSM struct {
	Name            string            `json:"name" yaml:"name"`
	WithDecryption  bool              `json:"withDecryption,omitempty" yaml:"withDecryption,omitempty"`
	Region          string            `json:"region,omitempty" yaml:"region,omitempty"`
	AccessKeyID     *ValueOrSecretRef `json:"accessKeyId,omitempty" yaml:"accessKeyId,omitempty"`
	SecretAccessKey *ValueOrSecretRef `json:"secretAccessKey,omitempty" yaml:"secretAccessKey,omitempty"`
}
