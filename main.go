/*
 * Copyright 2025 PKA-OpenLD
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logger.WithField("config", *configFile).Info("Starting PMMNM MQTT Bridge")

	// Load configuration
	config, err := LoadConfig(*configFile)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		logger.WithError(err).Fatal("Invalid configuration")
	}

	// Configure logger based on config
	level, err := logrus.ParseLevel(config.Logging.Level)
	if err != nil {
		logger.WithError(err).Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	if config.Logging.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	logger.WithFields(logrus.Fields{
		"api_endpoint": config.API.Endpoint,
		"mqtt_broker":  config.MQTT.Broker,
		"topics":       len(config.Topics),
		"log_level":    config.Logging.Level,
	}).Info("Configuration loaded successfully")

	// Create API client
	apiClient := NewAPIClient(config.API.Endpoint, config.API.Timeout, logger)

	// Create MQTT bridge
	bridge := NewMQTTBridge(config, apiClient, logger)

	// Connect to MQTT broker
	if err := bridge.Connect(); err != nil {
		logger.WithError(err).Fatal("Failed to connect to MQTT broker")
	}

	logger.Info("Bridge is running. Press Ctrl+C to exit.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down bridge...")
	bridge.Disconnect()
	logger.Info("Bridge stopped")
}
