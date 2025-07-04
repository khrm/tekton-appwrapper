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
	"context"

	"github.com/tektoncd/tekton-appwrapper/pkg/config"
	"knative.dev/pkg/configmap"

	pipelineruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1/pipelinerun"
	pipelinerunreconciler "github.com/tektoncd/pipeline/pkg/client/injection/reconciler/pipeline/v1/pipelinerun"
	"k8s.io/client-go/dynamic"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/logging"
)

// NewController creates a Controller for watching PipelineRuns.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	pipelineRunInformer := pipelineruninformer.Get(ctx)
	pipelineRunLister := pipelineRunInformer.Lister()
	logger := logging.FromContext(ctx)
	configStore := config.NewStore(logger.Named("config-store"))
	configStore.WatchConfigs(cmw)

	// Get rest.Config and create dynamic client
	k8scfg := injection.GetConfig(ctx)
	dynamicClient := dynamic.NewForConfigOrDie(k8scfg)

	c := &Reconciler{
		pipelineRunLister: pipelineRunLister,
		dynamicClient:     dynamicClient,
	}

	impl := pipelinerunreconciler.NewImpl(ctx, c, func(_ *controller.Impl) controller.Options {
		return controller.Options{
			// This appwrapper pipelinerun reconciler shouldn't mutate the pipelinerun's status.
			SkipStatusUpdates: true,
			ConfigStore:       configStore,
		}
	})

	// The config store is already set up to watch for changes above

	_, err := pipelineRunInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))
	if err != nil {
		logger.Panicf("Couldn't register PipelineRun informer event handler: %w", err)
	}

	return impl
}
