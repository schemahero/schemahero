package slackapp

import (
	"context"
	stderrors "errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	schemasclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SlackChannelAnnotation     = "replicated.com/slack-channel"
	SlackMessageTSAnnotation   = "replicated.com/slack-message-ts"
	SlackInvalidReactionUsers  = "replicated.com/slack-invalid-reaction-users"
	SlackApprovalReplyUser     = "replicated.com/slack-approval-reply-user"
	SlackRejectionReplyUser    = "replicated.com/slack-rejection-reply-user"
	ApproveReaction            = "white_check_mark"
	DenyReaction               = "x"
	DefaultPollInterval        = 15 * time.Second
	DefaultPendingPollInterval = 5 * time.Second
)

type Poller struct {
	schemasClient   schemasclientv1alpha4.SchemasV1alpha4Interface
	slackClient     SlackClient
	dispatcher      Dispatcher
	channel         string
	namespace       string
	pollInterval    time.Duration
	pendingInterval time.Duration
	dispatchTimeout time.Duration
	dispatched      map[string]bool
	dispatchedMu    sync.Mutex
	now             func() time.Time
}

type syncResult struct {
	pending    bool
	retryAfter time.Duration
}

func NewPoller(schemasClient schemasclientv1alpha4.SchemasV1alpha4Interface, slackClient SlackClient, channel string, namespace string) *Poller {
	return &Poller{
		schemasClient:   schemasClient,
		slackClient:     slackClient,
		channel:         channel,
		namespace:       namespace,
		pollInterval:    DefaultPollInterval,
		pendingInterval: DefaultPendingPollInterval,
		dispatchTimeout: DefaultDispatchTimeout,
		dispatched:      map[string]bool{},
		now:             time.Now,
	}
}

func (p *Poller) WithDispatcher(dispatcher Dispatcher) *Poller {
	p.dispatcher = dispatcher
	return p
}

func (p *Poller) Run(ctx context.Context) error {
	if p.channel == "" {
		return errors.New("slack channel is required")
	}
	if p.pollInterval == 0 {
		p.pollInterval = DefaultPollInterval
	}
	if p.pendingInterval == 0 {
		p.pendingInterval = DefaultPendingPollInterval
	}

	for {
		interval := p.syncAndLog(ctx)
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (p *Poller) syncAndLog(ctx context.Context) time.Duration {
	result, err := p.sync(ctx)
	if err != nil {
		log.Printf("schemahero slack app sync failed: %v", err)
		if result.retryAfter > 0 {
			return result.retryAfter
		}
	}
	if result.pending {
		return p.pendingInterval
	}
	return p.pollInterval
}

func (p *Poller) Sync(ctx context.Context) error {
	_, err := p.sync(ctx)
	return err
}

func (p *Poller) sync(ctx context.Context) (syncResult, error) {
	result := syncResult{}
	migrations, err := p.schemasClient.Migrations(p.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return result, errors.Wrap(err, "failed to list migrations")
	}

	var syncErr error
	for i := range migrations.Items {
		migration := &migrations.Items[i]
		if migration.Status.Phase != schemasv1alpha4.Planned {
			continue
		}
		result.pending = true

		updatedMessage, err := p.ensureApprovalMessage(ctx, migration)
		if err != nil {
			if result.retryAfter == 0 {
				result.retryAfter = retryAfter(err)
			}
			syncErr = stderrors.Join(syncErr, errors.Wrapf(err, "failed to ensure approval message for migration %s/%s", migration.Namespace, migration.Name))
			continue
		}
		if updatedMessage {
			continue
		}
		if err := p.handleReactions(ctx, migration); err != nil {
			if result.retryAfter == 0 {
				result.retryAfter = retryAfter(err)
			}
			syncErr = stderrors.Join(syncErr, errors.Wrapf(err, "failed to handle reactions for migration %s/%s", migration.Namespace, migration.Name))
			continue
		}
	}

	return result, syncErr
}

func retryAfter(err error) time.Duration {
	var rateLimitErr *RateLimitError
	if stderrors.As(err, &rateLimitErr) {
		return rateLimitErr.RetryAfter
	}
	return 0
}

func (p *Poller) ensureApprovalMessage(ctx context.Context, migration *schemasv1alpha4.Migration) (bool, error) {
	annotations := migration.GetAnnotations()
	planHash := schemasv1alpha4.PlanHashForDDL(migration.Spec.GeneratedDDL)
	if annotations[SlackMessageTSAnnotation] != "" && migration.Status.PlanHash == planHash {
		return false, nil
	}

	text := fmt.Sprintf("SchemaHero migration approval required for %s/%s", migration.Namespace, migration.Name)
	blocks := []map[string]interface{}{
		{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*SchemaHero migration approval required*\nMigration: `%s/%s`\nPlan hash: `%s`\nReact with :white_check_mark: to approve or :x: to deny.", migration.Namespace, migration.Name, planHash),
			},
		},
		{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": "```" + migration.Spec.GeneratedDDL + "```",
			},
		},
	}
	ts, err := p.slackClient.PostMessage(ctx, p.channel, blocks, text, "")
	if err != nil {
		return false, errors.Wrap(err, "failed to post slack approval message")
	}

	updated := migration.DeepCopy()
	annotations = updated.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[SlackChannelAnnotation] = p.channel
	annotations[SlackMessageTSAnnotation] = ts
	delete(annotations, SlackInvalidReactionUsers)
	delete(annotations, SlackApprovalReplyUser)
	delete(annotations, SlackRejectionReplyUser)
	updated.SetAnnotations(annotations)
	updated.Status.PlanHash = planHash
	_, err = p.schemasClient.Migrations(updated.Namespace).Update(ctx, updated, metav1.UpdateOptions{})
	return true, errors.Wrap(err, "failed to update migration slack approval annotations")
}

func (p *Poller) handleReactions(ctx context.Context, migration *schemasv1alpha4.Migration) error {
	annotations := migration.GetAnnotations()
	channel := annotations[SlackChannelAnnotation]
	ts := annotations[SlackMessageTSAnnotation]
	if channel == "" || ts == "" {
		return nil
	}

	reactions, err := p.slackClient.Reactions(ctx, channel, ts)
	if err != nil {
		return errors.Wrap(err, "failed to fetch slack reactions")
	}

	for _, reaction := range reactions {
		switch reaction.Name {
		case ApproveReaction:
			if len(reaction.Users) == 0 {
				continue
			}
			return p.approve(ctx, migration, reaction.Users[0])
		case DenyReaction:
			if len(reaction.Users) == 0 {
				continue
			}
			return p.reject(ctx, migration, reaction.Users[0])
		default:
			if err := p.warnInvalidReaction(ctx, migration, reaction.Name, reaction.Users); err != nil {
				return errors.Wrapf(err, "failed to warn users for invalid slack reaction %q", reaction.Name)
			}
		}
	}

	return nil
}

func (p *Poller) approve(ctx context.Context, migration *schemasv1alpha4.Migration, userID string) error {
	actor, err := p.slackClient.UserName(ctx, userID)
	if err != nil {
		actor = userID
	}
	planHash := migration.Status.PlanHash
	if planHash == "" {
		planHash = schemasv1alpha4.PlanHashForDDL(migration.Spec.GeneratedDDL)
	}

	updated := migration.DeepCopy()
	updated.Status.ApprovedAt = p.now().Unix()
	updated.Status.ApprovedBy = actor
	updated.Status.ApprovedPlanHash = planHash
	updated.Status.PlanHash = planHash
	updated.Status.Phase = schemasv1alpha4.Approved
	p.setAuditReplyAnnotation(updated, SlackApprovalReplyUser, userID)
	updatedMigration, err := p.schemasClient.Migrations(updated.Namespace).Update(ctx, updated, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to update approved migration")
	}
	if err := p.postAuditReply(ctx, migration, SlackApprovalReplyUser, userID, fmt.Sprintf("<@%s> approved this migration at %s.", userID, p.now().UTC().Format(time.RFC3339))); err != nil {
		return errors.Wrap(err, "failed to post slack approval audit reply")
	}

	if release, ok := releaseFromMigration(updatedMigration); ok && p.dispatcher != nil {
		go p.dispatchWhenReleaseExecuted(ctx, release, updatedMigration.Namespace)
	}
	return nil
}

func (p *Poller) reject(ctx context.Context, migration *schemasv1alpha4.Migration, userID string) error {
	actor, err := p.slackClient.UserName(ctx, userID)
	if err != nil {
		actor = userID
	}

	updated := migration.DeepCopy()
	updated.Status.RejectedAt = p.now().Unix()
	updated.Status.RejectedBy = actor
	updated.Status.Phase = schemasv1alpha4.Rejected
	p.setAuditReplyAnnotation(updated, SlackRejectionReplyUser, userID)
	_, err = p.schemasClient.Migrations(updated.Namespace).Update(ctx, updated, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to update rejected migration")
	}
	return errors.Wrap(p.postAuditReply(ctx, migration, SlackRejectionReplyUser, userID, fmt.Sprintf("<@%s> denied this migration at %s.", userID, p.now().UTC().Format(time.RFC3339))), "failed to post slack rejection audit reply")
}

func (p *Poller) setAuditReplyAnnotation(migration *schemasv1alpha4.Migration, annotation string, userID string) {
	annotations := migration.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[annotation] = userID
	migration.SetAnnotations(annotations)
}

func (p *Poller) postAuditReply(ctx context.Context, migration *schemasv1alpha4.Migration, annotation string, userID string, text string) error {
	annotations := migration.GetAnnotations()
	if annotations[annotation] == userID {
		return nil
	}
	channel := annotations[SlackChannelAnnotation]
	ts := annotations[SlackMessageTSAnnotation]
	if channel == "" || ts == "" {
		return nil
	}
	_, err := p.slackClient.PostMessage(ctx, channel, nil, text, ts)
	return errors.Wrap(err, "failed to post slack thread reply")
}

func (p *Poller) warnInvalidReaction(ctx context.Context, migration *schemasv1alpha4.Migration, reaction string, users []string) error {
	if len(users) == 0 {
		return nil
	}

	annotations := migration.GetAnnotations()
	alreadyWarned := strings.Split(annotations[SlackInvalidReactionUsers], ",")
	warned := map[string]bool{}
	for _, user := range alreadyWarned {
		if user != "" {
			warned[user] = true
		}
	}

	newWarnings := []string{}
	for _, user := range users {
		key := reaction + ":" + user
		if warned[key] {
			continue
		}
		newWarnings = append(newWarnings, key)
		_, err := p.slackClient.PostMessage(
			ctx,
			annotations[SlackChannelAnnotation],
			nil,
			fmt.Sprintf("<@%s> `%s` is not a valid approval reaction. Use :white_check_mark: to approve or :x: to deny.", user, reaction),
			annotations[SlackMessageTSAnnotation],
		)
		if err != nil {
			return errors.Wrap(err, "failed to post invalid reaction warning")
		}
	}

	if len(newWarnings) == 0 {
		return nil
	}

	updated := migration.DeepCopy()
	annotations = updated.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	if annotations[SlackInvalidReactionUsers] != "" {
		annotations[SlackInvalidReactionUsers] += ","
	}
	annotations[SlackInvalidReactionUsers] += strings.Join(newWarnings, ",")
	updated.SetAnnotations(annotations)
	_, err := p.schemasClient.Migrations(updated.Namespace).Update(ctx, updated, metav1.UpdateOptions{})
	return errors.Wrap(err, "failed to update invalid reaction warning annotations")
}

func (p *Poller) dispatchWhenReleaseExecuted(ctx context.Context, release Release, namespace string) {
	timeout := p.dispatchTimeout
	if timeout == 0 {
		timeout = DefaultDispatchTimeout
	}
	pollInterval := p.pollInterval
	if pollInterval == 0 {
		pollInterval = DefaultPollInterval
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		ready, err := releaseMigrationsExecuted(ctx, p.schemasClient, release, namespace)
		if err != nil {
			log.Printf("schemahero slack app release dispatch readiness failed: %v", err)
			if releaseReadinessErrorIsTerminal(err) {
				return
			}
		} else if ready {
			if p.markDispatched(release.DispatchKey()) {
				_ = p.dispatcher.Dispatch(ctx, release)
			}
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (p *Poller) markDispatched(key string) bool {
	p.dispatchedMu.Lock()
	defer p.dispatchedMu.Unlock()
	if p.dispatched[key] {
		return false
	}
	p.dispatched[key] = true
	return true
}
