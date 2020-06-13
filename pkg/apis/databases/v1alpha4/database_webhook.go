/*

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
// +kubebuilder:docs-gen:collapse=Apache License

package v1alpha4

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	validationutils "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:docs-gen:collapse=Go imports

/*
This setup is doubles as setup for our conversion webhooks: as long as our
types implement the
[Hub](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/conversion#Hub) and
[Convertible](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/conversion#Convertible)
interfaces, a conversion webhook will be registered.
*/

func (r *Database) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

/*
 */
var _ webhook.Defaulter = &Database{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Database) Default() {

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-batch-databases-schemahero-io-v1alpha4-database,mutating=false,failurePolicy=fail,groups=databases.schemahero.io,resources=databases,versions=v1alpha4,name=databases.databases.schemmahero.io
var _ webhook.Validator = &Database{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateCreate() error {
	return r.validateDatabase()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateUpdate(old runtime.Object) error {
	return r.validateDatabase()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateDelete() error {
	return nil
}

func (r *Database) validateDatabase() error {
	var allErrs field.ErrorList
	if err := r.validateDatabaseName(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateDatabaseSpec(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "databases.schemahero.io", Kind: "Database"},
		r.Name, allErrs)
}

func (r *Database) validateDatabaseSpec() *field.Error {
	return nil
}

func (r *Database) validateDatabaseName() *field.Error {
	if r.ObjectMeta.Name == "airlinedb" {
		return field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "must be not set to airlinedb")
	}

	if len(r.ObjectMeta.Name) > validationutils.DNS1035LabelMaxLength-11 {
		// The job name length is 63 character like all Kubernetes objects
		// (which must fit in a DNS subdomain). The cronjob controller appends
		// a 11-character suffix to the cronjob (`-$TIMESTAMP`) when creating
		// a job. The job name length limit is 63 characters. Therefore cronjob
		// names must have length <= 63-11=52. If we don't validate this here,
		// then job creation will fail later.
		return field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "must be no more than 52 characters")
	}
	return nil
}

// +kubebuilder:docs-gen:collapse=Existing Defaulting and Validation
