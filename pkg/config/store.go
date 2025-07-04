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

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/configmap"
)

const (
	// ConfigMapName is the name of the ConfigMap that the appwrapper controller
	// expects to be populated with the configuration.
	ConfigMapName = "tekton-appwrapper-config"

	// DefaultQueueNameKey is the key in the ConfigMap for the default queue name
	DefaultQueueNameKey = "default-queue"
)

// Store is a typed wrapper around configmap.Untyped store to handle our configmaps.
type Store struct {
	*configmap.UntypedStore
}

// NewStore creates a new store of Configs and optionally calls functions when ConfigMaps are updated.
func NewStore(logger configmap.Logger, onAfterStore ...func(name string, value any)) *Store {
	store := &Store{
		UntypedStore: configmap.NewUntypedStore(
			"appwrapper",
			logger,
			configmap.Constructors{
				ConfigMapName: NewConfigFromConfigMap,
			},
			onAfterStore...,
		),
	}

	return store
}

// ToContext attaches the current Config state to the provided context.
func (s *Store) ToContext(ctx context.Context) context.Context {
	return ToContext(ctx, s.Load())
}

// Load creates a Config from the current config state of the Store.
func (s *Store) Load() *Config {
	if s.UntypedStore != nil {
		if cfg := s.UntypedStore.UntypedLoad(ConfigMapName); cfg != nil {
			if typedCfg, ok := cfg.(*Config); ok {
				return typedCfg
			}
		}
	}

	// Return default config if nothing is loaded
	return &Config{DefaultQueueName: "default-queue"}
}

// NewConfigFromConfigMap creates a Config from the supplied ConfigMap
func NewConfigFromConfigMap(config *corev1.ConfigMap) (*Config, error) {
	cfg := &Config{}

	if defaultQueueName, ok := config.Data[DefaultQueueNameKey]; ok {
		cfg.DefaultQueueName = defaultQueueName
	} else {
		cfg.DefaultQueueName = "default-queue" // fallback default
	}

	return cfg, nil
}
