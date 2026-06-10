package slackapp

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	ReleaseTagAnnotation        = "replicated.com/release-tag"
	ReleaseSHAAnnotation        = "replicated.com/release-sha"
	PreviousSHAAnnotation       = "replicated.com/previous-sha"
	PreviousTagNameAnnotation   = "replicated.com/previous-tag-name"
	DefaultDispatchPollInterval = 10 * time.Second
	DefaultDispatchTimeout      = 30 * time.Minute
	defaultDepotBinary          = "depot"
	productionReleaseTagNeedle  = `release_tag="${PRODUCTION_RELEASE_TAG:-${{ inputs.tag || github.ref_name }}}"`
	productionReleaseSHANeedle  = `release_sha="${PRODUCTION_RELEASE_SHA:-${{ inputs.sha || github.sha }}}"`
	productionPreviousSHANeedle = `previous_sha="${PRODUCTION_PREVIOUS_SHA:-${{ inputs.previous-sha }}}"`
	productionPreviousTagNeedle = `previous_tag_name="${PRODUCTION_PREVIOUS_TAG_NAME:-${{ inputs.previous-tag-name }}}"`
)

type Release struct {
	Tag             string
	SHA             string
	PreviousSHA     string
	PreviousTagName string
}

func (r Release) DispatchKey() string {
	return r.Tag + ":" + r.SHA
}

type Dispatcher interface {
	Dispatch(ctx context.Context, release Release) error
}

type DepotDispatcher struct {
	Token                string
	Repo                 string
	WorkflowTemplatePath string
	DepotBinary          string
}

func NewDepotDispatcherFromEnv() (*DepotDispatcher, bool) {
	token := os.Getenv("DEPOT_CI_DISPATCH_TOKEN")
	workflowTemplatePath := os.Getenv("DEPOT_WORKFLOW_TEMPLATE_PATH")
	repo := os.Getenv("DEPOT_REPO")
	if token == "" || workflowTemplatePath == "" || repo == "" {
		return nil, false
	}

	depotBinary := os.Getenv("DEPOT_BINARY")
	if depotBinary == "" {
		depotBinary = defaultDepotBinary
	}

	return &DepotDispatcher{
		Token:                token,
		Repo:                 repo,
		WorkflowTemplatePath: workflowTemplatePath,
		DepotBinary:          depotBinary,
	}, true
}

func (d *DepotDispatcher) Dispatch(ctx context.Context, release Release) error {
	if d.Token == "" {
		return errors.New("depot token is required")
	}
	if d.WorkflowTemplatePath == "" {
		return errors.New("depot workflow template path is required")
	}
	if d.Repo == "" {
		return errors.New("depot repo is required")
	}

	renderedWorkflow, err := renderWorkflowTemplate(d.WorkflowTemplatePath, release)
	if err != nil {
		return err
	}
	defer os.Remove(renderedWorkflow)

	depotBinary := d.DepotBinary
	if depotBinary == "" {
		depotBinary = defaultDepotBinary
	}
	cmd := exec.CommandContext(ctx, depotBinary,
		"ci", "run",
		"--repo", d.Repo,
		"--workflow", renderedWorkflow,
		"--token", d.Token,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "failed to dispatch depot workflow: %s", string(out))
	}
	return nil
}

func renderWorkflowTemplate(templatePath string, release Release) (string, error) {
	workflowBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}

	workflow := string(workflowBytes)
	replacements := map[string]string{
		productionReleaseTagNeedle:  `release_tag="` + release.Tag + `"`,
		productionReleaseSHANeedle:  `release_sha="` + release.SHA + `"`,
		productionPreviousSHANeedle: `previous_sha="` + release.PreviousSHA + `"`,
		productionPreviousTagNeedle: `previous_tag_name="` + release.PreviousTagName + `"`,
	}

	for needle, replacement := range replacements {
		if !strings.Contains(workflow, needle) {
			return "", errors.Errorf("workflow template does not contain expected marker: %s", needle)
		}
		workflow = strings.ReplaceAll(workflow, needle, replacement)
	}

	rendered, err := os.CreateTemp("", "schemahero-production-continuation-*.yml")
	if err != nil {
		return "", err
	}
	defer rendered.Close()

	if _, err := rendered.WriteString(workflow); err != nil {
		os.Remove(rendered.Name())
		return "", err
	}

	return rendered.Name(), nil
}
