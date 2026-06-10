package slackapp

import (
	"context"
	stderrors "errors"
	"strings"
	"testing"
	"time"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	testclient "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/fake"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeSlackClient struct {
	messages     []fakeSlackMessage
	reactions    []Reaction
	reactionsErr error
	users        map[string]string
	failPostText string
}

type fakeSlackMessage struct {
	channel  string
	text     string
	threadTS string
}

func (c *fakeSlackClient) PostMessage(_ context.Context, channel string, _ []map[string]interface{}, text string, threadTS string) (string, error) {
	if c.failPostText != "" && strings.Contains(text, c.failPostText) {
		return "", stderrors.New("post failed")
	}
	c.messages = append(c.messages, fakeSlackMessage{
		channel:  channel,
		text:     text,
		threadTS: threadTS,
	})
	return "123.456", nil
}

func (c *fakeSlackClient) Reactions(_ context.Context, _ string, _ string) ([]Reaction, error) {
	if c.reactionsErr != nil {
		return nil, c.reactionsErr
	}
	return c.reactions, nil
}

func (c *fakeSlackClient) UserName(_ context.Context, userID string) (string, error) {
	if c.users != nil && c.users[userID] != "" {
		return c.users[userID], nil
	}
	return userID, nil
}

func TestPollerPostsApprovalMessage(t *testing.T) {
	migration := plannedMigration("approval-rehearsal")
	client := testclient.NewSimpleClientset(migration).SchemasV1alpha4()
	slack := &fakeSlackClient{}
	poller := NewPoller(client, slack, "C123", "default")

	require.NoError(t, poller.Sync(context.Background()))

	require.Len(t, slack.messages, 1)
	require.Contains(t, slack.messages[0].text, "approval-rehearsal")
	updated, err := client.Migrations("default").Get(context.Background(), "approval-rehearsal", metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, "C123", updated.Annotations[SlackChannelAnnotation])
	require.Equal(t, "123.456", updated.Annotations[SlackMessageTSAnnotation])
}

func TestPollerRepostsApprovalMessageWhenPlanHashChanges(t *testing.T) {
	migration := plannedMigration("approval-rehearsal")
	migration.Spec.GeneratedDDL = "select 2"
	migration.Status.PlanHash = "old-plan"
	migration.Annotations[SlackChannelAnnotation] = "C123"
	migration.Annotations[SlackMessageTSAnnotation] = "old-ts"
	migration.Annotations[SlackInvalidReactionUsers] = "eyes:U123"
	client := testclient.NewSimpleClientset(migration).SchemasV1alpha4()
	slack := &fakeSlackClient{
		reactions: []Reaction{{Name: ApproveReaction, Users: []string{"U123"}}},
	}
	poller := NewPoller(client, slack, "C123", "default")

	require.NoError(t, poller.Sync(context.Background()))

	require.Len(t, slack.messages, 1)
	updated, err := client.Migrations("default").Get(context.Background(), "approval-rehearsal", metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, schemasv1alpha4.Planned, updated.Status.Phase)
	require.Equal(t, schemasv1alpha4.PlanHashForDDL("select 2"), updated.Status.PlanHash)
	require.Equal(t, "123.456", updated.Annotations[SlackMessageTSAnnotation])
	require.NotContains(t, updated.Annotations, SlackInvalidReactionUsers)
}

func TestPollerContinuesAfterMigrationError(t *testing.T) {
	broken := plannedMigration("broken")
	ok := plannedMigration("ok")
	client := testclient.NewSimpleClientset(broken, ok).SchemasV1alpha4()
	slack := &fakeSlackClient{failPostText: "broken"}
	poller := NewPoller(client, slack, "C123", "default")

	err := poller.Sync(context.Background())

	require.Error(t, err)
	require.Len(t, slack.messages, 1)
	require.Contains(t, slack.messages[0].text, "ok")
	updated, getErr := client.Migrations("default").Get(context.Background(), "ok", metav1.GetOptions{})
	require.NoError(t, getErr)
	require.Equal(t, "123.456", updated.Annotations[SlackMessageTSAnnotation])
}

func TestPollerUsesPendingIntervalWhenMigrationNeedsApproval(t *testing.T) {
	migration := plannedMigration("approval-rehearsal")
	client := testclient.NewSimpleClientset(migration).SchemasV1alpha4()
	slack := &fakeSlackClient{}
	poller := NewPoller(client, slack, "C123", "default")
	poller.pollInterval = 15 * time.Second
	poller.pendingInterval = 5 * time.Second

	require.Equal(t, 5*time.Second, poller.syncAndLog(context.Background()))
}

func TestPollerUsesRetryAfterWhenSlackRateLimited(t *testing.T) {
	migration := plannedMigration("approval-rehearsal")
	migration.Annotations[SlackChannelAnnotation] = "C123"
	migration.Annotations[SlackMessageTSAnnotation] = "123.456"
	client := testclient.NewSimpleClientset(migration).SchemasV1alpha4()
	slack := &fakeSlackClient{
		reactionsErr: &RateLimitError{RetryAfter: 30 * time.Second},
	}
	poller := NewPoller(client, slack, "C123", "default")
	poller.pollInterval = 15 * time.Second
	poller.pendingInterval = 5 * time.Second

	require.Equal(t, 30*time.Second, poller.syncAndLog(context.Background()))
}

func TestPollerApprovesOnWhiteCheckReaction(t *testing.T) {
	migration := plannedMigration("approval-rehearsal")
	migration.Annotations[SlackChannelAnnotation] = "C123"
	migration.Annotations[SlackMessageTSAnnotation] = "123.456"
	client := testclient.NewSimpleClientset(migration).SchemasV1alpha4()
	slack := &fakeSlackClient{
		reactions: []Reaction{{Name: ApproveReaction, Users: []string{"U123"}}},
		users:     map[string]string{"U123": "Marc"},
	}
	poller := NewPoller(client, slack, "C123", "default")
	poller.now = func() time.Time { return time.Unix(100, 0) }

	require.NoError(t, poller.Sync(context.Background()))

	updated, err := client.Migrations("default").Get(context.Background(), "approval-rehearsal", metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, schemasv1alpha4.Approved, updated.Status.Phase)
	require.Equal(t, "Marc", updated.Status.ApprovedBy)
	require.Equal(t, updated.Status.PlanHash, updated.Status.ApprovedPlanHash)
	require.Len(t, slack.messages, 1)
	require.Equal(t, "123.456", slack.messages[0].threadTS)
	require.Contains(t, slack.messages[0].text, "<@U123> approved this migration at 1970-01-01T00:01:40Z.")
}

func TestPollerRejectsOnXReaction(t *testing.T) {
	migration := plannedMigration("approval-rehearsal")
	migration.Annotations[SlackChannelAnnotation] = "C123"
	migration.Annotations[SlackMessageTSAnnotation] = "123.456"
	client := testclient.NewSimpleClientset(migration).SchemasV1alpha4()
	slack := &fakeSlackClient{
		reactions: []Reaction{{Name: DenyReaction, Users: []string{"U123"}}},
		users:     map[string]string{"U123": "Marc"},
	}
	poller := NewPoller(client, slack, "C123", "default")
	poller.now = func() time.Time { return time.Unix(100, 0) }

	require.NoError(t, poller.Sync(context.Background()))

	updated, err := client.Migrations("default").Get(context.Background(), "approval-rehearsal", metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, schemasv1alpha4.Rejected, updated.Status.Phase)
	require.Equal(t, "Marc", updated.Status.RejectedBy)
	require.Len(t, slack.messages, 1)
	require.Equal(t, "123.456", slack.messages[0].threadTS)
	require.Contains(t, slack.messages[0].text, "<@U123> denied this migration at 1970-01-01T00:01:40Z.")
}

func TestPollerWarnsOnInvalidReaction(t *testing.T) {
	migration := plannedMigration("approval-rehearsal")
	migration.Annotations[SlackChannelAnnotation] = "C123"
	migration.Annotations[SlackMessageTSAnnotation] = "123.456"
	client := testclient.NewSimpleClientset(migration).SchemasV1alpha4()
	slack := &fakeSlackClient{
		reactions: []Reaction{{Name: "eyes", Users: []string{"U123"}}},
	}
	poller := NewPoller(client, slack, "C123", "default")

	require.NoError(t, poller.Sync(context.Background()))

	require.Len(t, slack.messages, 1)
	require.Equal(t, "123.456", slack.messages[0].threadTS)
	require.Contains(t, slack.messages[0].text, "Use :white_check_mark: to approve or :x: to deny")
	updated, err := client.Migrations("default").Get(context.Background(), "approval-rehearsal", metav1.GetOptions{})
	require.NoError(t, err)
	require.Contains(t, updated.Annotations[SlackInvalidReactionUsers], "eyes:U123")
}

func plannedMigration(name string) *schemasv1alpha4.Migration {
	ddl := "select 1"
	return &schemasv1alpha4.Migration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Annotations: map[string]string{
				ReleaseTagAnnotation:      "v2099.01.01-0",
				ReleaseSHAAnnotation:      "1111111111111111111111111111111111111111",
				PreviousSHAAnnotation:     "0000000000000000000000000000000000000000",
				PreviousTagNameAnnotation: "v2098.12.31-0",
			},
		},
		Spec: schemasv1alpha4.MigrationSpec{
			GeneratedDDL: ddl,
		},
		Status: schemasv1alpha4.MigrationStatus{
			Phase:    schemasv1alpha4.Planned,
			PlanHash: schemasv1alpha4.PlanHashForDDL(ddl),
		},
	}
}
