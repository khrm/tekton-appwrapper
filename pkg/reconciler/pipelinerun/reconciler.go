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
	"fmt"

	appwrapperv1beta2 "github.com/project-codeflare/appwrapper/api/v1beta2"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	pipelinerunreconciler "github.com/tektoncd/pipeline/pkg/client/injection/reconciler/pipeline/v1/pipelinerun"
	v1 "github.com/tektoncd/pipeline/pkg/client/listers/pipeline/v1"
	"github.com/tektoncd/tekton-appwrapper/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"knative.dev/pkg/logging"
	knativereconciler "knative.dev/pkg/reconciler"
)

// We don't need this. It's just to create a podTemplateSpec for the AppWrapper.
const defaultContainerImage = "tektoncd.dev/tekton-appwrapper:latest"

// Reconciler reconciles a PipelineRun object
type Reconciler struct {
	pipelineRunLister v1.PipelineRunLister
	dynamicClient     dynamic.Interface
}

// Check that our Reconciler implements pipelinerunreconciler.Interface
var _ pipelinerunreconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements the business logic for reconciling PipelineRun objects
func (r *Reconciler) ReconcileKind(ctx context.Context, pipelineRun *tektonv1.PipelineRun) knativereconciler.Event {
	logger := logging.FromContext(ctx).With("pipelineRun", pipelineRun.Name)

	logger.Infof("Reconciling PipelineRun '%s/%s'", pipelineRun.Namespace, pipelineRun.Name)
	// Get the current configuration from context
	cfg := config.FromContext(ctx)
	if cfg == nil {
		logger.Infof("cfg is nil, using default config")
		cfg = &config.Config{} // fallback to default config
	}
	logger.Infof("cfg.DefaultQueueName(): %v", cfg.GetDefaultQueueName())

	logger.Infof("Initiating reconciliation for PipelineRun '%s/%s'", pipelineRun.Namespace, pipelineRun.Name)
	if pipelineRun.IsDone() {
		logger.Infof("PipelineRun %s is already completed, skipping AppWrapper creation", pipelineRun.Name)
		return nil
	}

	// Check if AppWrapper already exists for this PipelineRun
	existingAppWrapper, err := r.checkExistingAppWrapper(ctx, pipelineRun)
	if err != nil {
		logger.Errorf("Error checking for existing AppWrapper: %v", err)
		return knativereconciler.NewEvent(corev1.EventTypeWarning, "AppWrapperCheckFailed", "Failed to check for existing AppWrapper: %v", err)
	}

	if existingAppWrapper != nil {
		logger.Infof("AppWrapper already exists for PipelineRun %s, skipping creation", pipelineRun.Name)
		return nil
	}

	// Create AppWrapper
	appWrapper, err := r.createAppWrapper(ctx, pipelineRun, cfg)
	if err != nil {
		logger.Errorf("Error creating AppWrapper: %v", err)
		return knativereconciler.NewEvent(corev1.EventTypeWarning, "AppWrapperCreationFailed", "Failed to create AppWrapper: %v", err)
	}

	logger.Infof("Successfully created AppWrapper %s for PipelineRun %s", appWrapper.GetName(), pipelineRun.Name)
	return knativereconciler.NewEvent(corev1.EventTypeNormal, "AppWrapperCreated", "Successfully created AppWrapper %s", appWrapper.GetName())
}

// checkExistingAppWrapper checks if an AppWrapper already exists for the given PipelineRun.
func (r *Reconciler) checkExistingAppWrapper(ctx context.Context, pipelineRun *tektonv1.PipelineRun) (*unstructured.Unstructured, error) {
	gvr := appwrapperv1beta2.GroupVersion.WithResource("appwrappers")
	appWrapperName := fmt.Sprintf("%s-appwrapper", pipelineRun.Name)

	existingAppWrapper, err := r.dynamicClient.Resource(gvr).Namespace(pipelineRun.Namespace).Get(ctx, appWrapperName, metav1.GetOptions{})
	if err == nil {
		return existingAppWrapper, nil
	}
	if !errors.IsNotFound(err) {
		return nil, err
	}
	return nil, nil
}

// createAppWrapper creates a new AppWrapper for the given PipelineRun.
func (r *Reconciler) createAppWrapper(ctx context.Context, pipelineRun *tektonv1.PipelineRun, cfg *config.Config) (*unstructured.Unstructured, error) {
	gvr := appwrapperv1beta2.GroupVersion.WithResource("appwrappers")
	appWrapperName := fmt.Sprintf("%s-appwrapper", pipelineRun.Name)

	// Extract resource requests from PipelineRun annotations
	requests := map[string]string{}
	if cpu, ok := pipelineRun.Annotations["kueue.tekton.dev/requests-cpu"]; ok {
		requests["cpu"] = cpu
	}
	if memory, ok := pipelineRun.Annotations["kueue.tekton.dev/requests-memory"]; ok {
		requests["memory"] = memory
	}
	if storage, ok := pipelineRun.Annotations["kueue.tekton.dev/requests-storage"]; ok {
		requests["storage"] = storage
	}
	if ephemeralStorage, ok := pipelineRun.Annotations["kueue.tekton.dev/requests-ephemeral-storage"]; ok {
		requests["ephemeral-storage"] = ephemeralStorage
	}

	// Create a copy of the PipelineRun to modify
	pipelineRunCopy := pipelineRun.DeepCopy()

	// Convert PipelineRun to unstructured to add the appwrapper field
	unstructuredPR, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pipelineRunCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to convert PipelineRun to Unstructured: %v", err)
	}

	// Filter out unwanted metadata fields
	if metadata, ok := unstructuredPR["metadata"].(map[string]interface{}); ok {
		// Remove fields that should not be copied to the AppWrapper template
		delete(metadata, "generateName")
		delete(metadata, "managedFields")
		delete(metadata, "generation")
		delete(metadata, "uid")
		delete(metadata, "resourceVersion")
		delete(metadata, "creationTimestamp")
	}

	// Clean up null values from the unstructured object
	cleanupNullValues(unstructuredPR)

	// Add the appwrapper field to the spec
	if unstructuredPR["spec"] == nil {
		unstructuredPR["spec"] = map[string]interface{}{}
	}
	spec := unstructuredPR["spec"].(map[string]interface{})
	// Build the resources.requests map only with present values
	resourceRequests := map[string]interface{}{}
	for k, v := range requests {
		resourceRequests[k] = v
	}
	resourceRequests["tekton.dev/pipelinerun"] = "1"

	// Add appwrapper field with PodTemplateSpec
	appwrapperField := map[string]interface{}{
		"replicas": float64(1),
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{
					"name":  "appwrapper-container",
					"image": "registry.k8s.io/nginx-slim:0.27",
					"resources": map[string]interface{}{
						"requests": resourceRequests,
					},
				},
			},
		},
	}
	spec["appwrapper"] = appwrapperField

	// Build the AppWrapper as unstructured
	aw := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": appwrapperv1beta2.GroupVersion.String(),
			"kind":       "AppWrapper",
			"metadata": map[string]interface{}{
				"name":      appWrapperName,
				"namespace": pipelineRun.Namespace,
				"labels": map[string]interface{}{
					"kueue.x-k8s.io/queue-name": cfg.GetDefaultQueueName(),
				},
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion": tektonv1.SchemeGroupVersion.String(),
						"kind":       "PipelineRun",
						"name":       pipelineRun.Name,
						"uid":        string(pipelineRun.UID),
						"controller": true,
					},
				},
			},
			"spec": map[string]interface{}{
				"components": []interface{}{
					map[string]interface{}{
						"template": map[string]interface{}{
							"apiVersion": "tekton.dev/v1beta1",
							"kind":       "PipelineRun",
							"metadata":   unstructuredPR["metadata"],
							"spec":       unstructuredPR["spec"],
						},
						"podSets": []interface{}{
							map[string]interface{}{
								"path":     "template.spec.appwrapper",
								"replicas": float64(1),
							},
						},
					},
				},
			},
		},
	}

	_, err = r.dynamicClient.Resource(gvr).Namespace(pipelineRun.Namespace).Create(ctx, aw, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create AppWrapper: %v", err)
	}

	return aw, nil
}

// cleanupNullValues recursively removes null values from a map[string]interface{}
func cleanupNullValues(obj map[string]interface{}) {
	for key, value := range obj {
		switch v := value.(type) {
		case nil:
			// Remove null values
			delete(obj, key)
		case map[string]interface{}:
			// Recursively clean nested maps
			cleanupNullValues(v)
			// If the map becomes empty after cleanup, remove it
			if len(v) == 0 {
				delete(obj, key)
			}
		case []interface{}:
			// Clean up arrays
			for i := len(v) - 1; i >= 0; i-- {
				if v[i] == nil {
					// Remove null elements from arrays
					v = append(v[:i], v[i+1:]...)
				} else if nestedMap, ok := v[i].(map[string]interface{}); ok {
					// Recursively clean nested maps in arrays
					cleanupNullValues(nestedMap)
					if len(nestedMap) == 0 {
						v = append(v[:i], v[i+1:]...)
					}
				}
			}
			// Update the array in the map
			if len(v) == 0 {
				delete(obj, key)
			} else {
				obj[key] = v
			}
		}
	}
}
