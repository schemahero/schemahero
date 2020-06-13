/*
Copyright 2019 Replicated, Inc.

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

package main

import (
	"flag"
	"os"

	"github.com/schemahero/schemahero/pkg/apis"
	// databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	"github.com/schemahero/schemahero/pkg/controller"
	"github.com/schemahero/schemahero/pkg/logger"
	"github.com/schemahero/schemahero/pkg/version"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8088", "The address the metric endpoint binds to.")
	flag.Parse()

	logger.Infof("Starting schemahero version %+v", version.GetBuild())

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{MetricsBindAddress: metricsAddr})
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	// Add webhooks for API conversion
	// if err = (&databasesv1alpha4.Database{}).SetupWebhookWithManager(mgr); err != nil {
	// 	logger.Error(err)
	// 	os.Exit(1)
	// }

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
