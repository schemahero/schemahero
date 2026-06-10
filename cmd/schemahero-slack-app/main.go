package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	schemasclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/config"
	"github.com/schemahero/schemahero/pkg/slackapp"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	if slackBotToken == "" {
		log.Fatal("SLACK_BOT_TOKEN is required")
	}
	slackChannelID := os.Getenv("SLACK_CHANNEL_ID")
	if slackChannelID == "" {
		log.Fatal("SLACK_CHANNEL_ID is required")
	}
	namespace := os.Getenv("SCHEMAHERO_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	cfg, err := config.GetRESTConfig()
	if err != nil {
		log.Fatal(err)
	}
	schemasClient, err := schemasclientv1alpha4.NewForConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	poller := slackapp.NewPoller(
		schemasClient,
		slackapp.NewWebAPIClient(slackBotToken),
		slackChannelID,
		namespace,
	)

	if dispatcher, ok := slackapp.NewDepotDispatcherFromEnv(); ok {
		poller.WithDispatcher(dispatcher)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("starting schemahero slack app poller for namespace %s", namespace)
	if err := poller.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
