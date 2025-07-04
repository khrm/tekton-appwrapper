// Copyright 2024 The Tekton Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"context"
)

// Config is the configuration for the appwrapper controller
type Config struct {
	// DefaultQueueName is the default queue name to use for AppWrappers
	DefaultQueueName string
}

// cfgKey is the key for the Config in the context.
type cfgKey struct{}

// FromContext extracts a Config from the provided context.
func FromContext(ctx context.Context) *Config {
	x := ctx.Value(cfgKey{})
	if x == nil {
		return nil
	}
	return x.(*Config)
}

// ToContext attaches the provided Config to the provided context, returning the
// new context with the Config attached.
func ToContext(ctx context.Context, c *Config) context.Context {
	return context.WithValue(ctx, cfgKey{}, c)
}

// GetDefaultQueueName returns the default queue name to use for AppWrappers
func (c *Config) GetDefaultQueueName() string {
	if c == nil {
		return "default-queue"
	}
	return c.DefaultQueueName
}
