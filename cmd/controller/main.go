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

package main

import (
	"context"
	"flag"
	"time"

	"github.com/tektoncd/tekton-appwrapper/pkg/reconciler/pipelinerun"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
)

var (
	threadiness = flag.Int("threadiness", controller.DefaultThreadsPerController, "Number of threads (Go routines) allocated to each controller")
	qps         = flag.Float64("qps", float64(rest.DefaultQPS), "Kubernetes client QPS setting")
	burst       = flag.Int("burst", rest.DefaultBurst, "Kubernetes client Burst setting")
	namespace   = flag.String("namespace", corev1.NamespaceAll, "Should the controller only watch a single namespace, then this value needs to be set to the namespace name otherwise leave it empty.")

	// AppWrapper specific flags
	requeueInterval = flag.Duration("requeue_interval", 10*time.Minute, "How long the controller waits to reprocess keys on certain events")
)

func main() {
	flag.Parse()
	ctx := signals.NewContext()

	// Allow users to customize the number of workers used to process the
	// controller's workqueue.
	controller.DefaultThreadsPerController = *threadiness

	ctors := []injection.ControllerConstructor{
		func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
			return pipelinerun.NewController(ctx, cmw)
		},
	}

	// This parses flags and gets the REST config
	k8scfg := injection.ParseAndGetRESTConfigOrDie()

	// Configure QPS and burst settings - multiply by number of controllers like reference implementation
	if qps != nil {
		k8scfg.QPS = float32(*qps) * float32(len(ctors))
	}
	if burst != nil {
		k8scfg.Burst = *burst * len(ctors)
	}

	// The application will now use ConfigMaps for logging configuration
	// config-logging and config-observability ConfigMaps should be deployed

	sharedmain.MainWithConfig(injection.WithNamespaceScope(ctx, *namespace), "appwrapper-pipelinerun-controller", k8scfg, ctors...)
}
