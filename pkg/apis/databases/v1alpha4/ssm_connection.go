/*
Copyright 2019 The SchemaHero Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha4

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getSSMConnection returns the driver, the resolved value, and any error
func (d *Database) getSSMConnection(ctx context.Context, clientset *kubernetes.Clientset, driver string, valueOrValueFrom ValueOrValueFrom) (string, string, error) {
	region := valueOrValueFrom.ValueFrom.SSM.Region
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create aws config")
	}
	cfg.Region = region

	if valueOrValueFrom.ValueFrom.SSM.AccessKeyID != nil && valueOrValueFrom.ValueFrom.SSM.SecretAccessKey != nil {
		accessKeyID := ""
		if valueOrValueFrom.ValueFrom.SSM.AccessKeyID.Value != "" {
			accessKeyID = valueOrValueFrom.ValueFrom.SSM.AccessKeyID.Value
		} else if valueOrValueFrom.ValueFrom.SSM.AccessKeyID.ValueFrom.SecretKeyRef != nil {
			secret, err := clientset.CoreV1().Secrets(d.Namespace).Get(ctx, valueOrValueFrom.ValueFrom.SSM.AccessKeyID.ValueFrom.SecretKeyRef.Name, metav1.GetOptions{})
			if err != nil {
				return "", "", errors.Wrap(err, "failed to get access key secret")
			}
			accessKeyID = string(secret.Data[valueOrValueFrom.ValueFrom.SSM.AccessKeyID.ValueFrom.SecretKeyRef.Key])
		}

		secretAccessKey := ""
		if valueOrValueFrom.ValueFrom.SSM.SecretAccessKey.Value != "" {
			secretAccessKey = valueOrValueFrom.ValueFrom.SSM.SecretAccessKey.Value
		} else if valueOrValueFrom.ValueFrom.SSM.SecretAccessKey.ValueFrom.SecretKeyRef != nil {
			secret, err := clientset.CoreV1().Secrets(d.Namespace).Get(ctx, valueOrValueFrom.ValueFrom.SSM.SecretAccessKey.ValueFrom.SecretKeyRef.Name, metav1.GetOptions{})
			if err != nil {
				return "", "", errors.Wrap(err, "failed to get secret access key secret")
			}
			accessKeyID = string(secret.Data[valueOrValueFrom.ValueFrom.SSM.SecretAccessKey.ValueFrom.SecretKeyRef.Key])
		}

		cfg.Credentials = credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")
	}

	client := ssm.NewFromConfig(cfg)

	params := ssm.GetParameterInput{
		WithDecryption: &valueOrValueFrom.ValueFrom.SSM.WithDecryption,
		Name:           aws.String(valueOrValueFrom.ValueFrom.SSM.Name),
	}
	out, err := client.GetParameter(ctx, &params)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get ssm parameter")
	}

	return driver, *out.Parameter.Value, nil
}
