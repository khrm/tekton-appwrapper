/*
Copyright 2024 IBM Corporation.

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

package pipelinerun

import (
	"time"
)

// Config defines shared reconciler configuration options for AppWrapper creation.
type Config struct {
	// How long the controller waits to reprocess keys on certain events
	RequeueInterval time.Duration

	// Whether to skip AppWrapper creation for completed PipelineRuns
	SkipCompletedPipelineRuns bool

	// Default queue name for AppWrapper resources
	DefaultQueueName string

	// Default container image for AppWrapper resources
	DefaultContainerImage string
}

// GetRequeueInterval returns the requeue interval for the controller.
func (c *Config) GetRequeueInterval() time.Duration {
	if c == nil {
		return 10 * time.Minute // Default value
	}
	return c.RequeueInterval
}

// GetSkipCompletedPipelineRuns returns whether to skip AppWrapper creation for completed PipelineRuns.
func (c *Config) GetSkipCompletedPipelineRuns() bool {
	if c == nil {
		return true // Default to skipping completed runs
	}
	return c.SkipCompletedPipelineRuns
}

// GetDefaultQueueName returns the default queue name for AppWrapper resources.
func (c *Config) GetDefaultQueueName() string {
	if c == nil || c.DefaultQueueName == "" {
		return "default-queue"
	}
	return c.DefaultQueueName
}

// GetDefaultContainerImage returns the default container image for AppWrapper resources.
func (c *Config) GetDefaultContainerImage() string {
	if c == nil || c.DefaultContainerImage == "" {
		return "registry.k8s.io/nginx-slim:0.27"
	}
	return c.DefaultContainerImage
}
