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
	"testing"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	v1 "github.com/tektoncd/pipeline/pkg/client/listers/pipeline/v1"
	"github.com/tektoncd/tekton-appwrapper/pkg/config"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"knative.dev/pkg/apis"
)

// Mock implementations
type mockPipelineRunLister struct {
	pipelineRuns map[string]*pipelinev1.PipelineRun
	errors       map[string]error
}

func (m *mockPipelineRunLister) PipelineRuns(namespace string) v1.PipelineRunNamespaceLister {
	return &mockPipelineRunNamespaceLister{
		pipelineRuns: m.pipelineRuns,
		errors:       m.errors,
		namespace:    namespace,
	}
}

func (m *mockPipelineRunLister) List(selector labels.Selector) (ret []*pipelinev1.PipelineRun, err error) {
	// Not used in current implementation
	return nil, nil
}

type mockPipelineRunNamespaceLister struct {
	pipelineRuns map[string]*pipelinev1.PipelineRun
	errors       map[string]error
	namespace    string
}

func (m *mockPipelineRunNamespaceLister) Get(name string) (*pipelinev1.PipelineRun, error) {
	key := fmt.Sprintf("%s/%s", m.namespace, name)
	if err, exists := m.errors[key]; exists {
		return nil, err
	}
	if pr, exists := m.pipelineRuns[key]; exists {
		return pr, nil
	}
	return nil, k8serrors.NewNotFound(schema.GroupResource{Group: "tekton.dev", Resource: "pipelineruns"}, name)
}

func (m *mockPipelineRunNamespaceLister) List(selector labels.Selector) (ret []*pipelinev1.PipelineRun, err error) {
	// Not used in current implementation
	return nil, nil
}

type mockDynamicClient struct {
	created map[string]*unstructured.Unstructured
	errors  map[string]error
}

func (m *mockDynamicClient) Resource(gvr schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return &mockNamespaceableResourceInterface{
		created: m.created,
		errors:  m.errors,
		gvr:     gvr,
	}
}

type mockNamespaceableResourceInterface struct {
	created map[string]*unstructured.Unstructured
	errors  map[string]error
	gvr     schema.GroupVersionResource
}

func (m *mockNamespaceableResourceInterface) Namespace(namespace string) dynamic.ResourceInterface {
	return &mockResourceInterface{
		created:   m.created,
		errors:    m.errors,
		gvr:       m.gvr,
		namespace: namespace,
	}
}

func (m *mockNamespaceableResourceInterface) Create(ctx context.Context, obj *unstructured.Unstructured, opts metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockNamespaceableResourceInterface) Update(ctx context.Context, obj *unstructured.Unstructured, opts metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockNamespaceableResourceInterface) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, opts metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockNamespaceableResourceInterface) Delete(ctx context.Context, name string, opts metav1.DeleteOptions, subresources ...string) error {
	// Not used in current implementation
	return nil
}

func (m *mockNamespaceableResourceInterface) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	// Not used in current implementation
	return nil
}

func (m *mockNamespaceableResourceInterface) Get(ctx context.Context, name string, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockNamespaceableResourceInterface) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockNamespaceableResourceInterface) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockNamespaceableResourceInterface) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockNamespaceableResourceInterface) Apply(ctx context.Context, name string, obj *unstructured.Unstructured, opts metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockNamespaceableResourceInterface) ApplyStatus(ctx context.Context, name string, obj *unstructured.Unstructured, opts metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	// Not used in current implementation
	return nil, nil
}

type mockResourceInterface struct {
	created   map[string]*unstructured.Unstructured
	errors    map[string]error
	gvr       schema.GroupVersionResource
	namespace string
}

func (m *mockResourceInterface) Create(ctx context.Context, obj *unstructured.Unstructured, opts metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	key := fmt.Sprintf("%s/%s", m.namespace, obj.GetName())
	if err, exists := m.errors[key]; exists {
		return nil, err
	}
	m.created[key] = obj
	return obj, nil
}

func (m *mockResourceInterface) Update(ctx context.Context, obj *unstructured.Unstructured, opts metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockResourceInterface) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, opts metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockResourceInterface) Delete(ctx context.Context, name string, opts metav1.DeleteOptions, subresources ...string) error {
	// Not used in current implementation
	return nil
}

func (m *mockResourceInterface) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	// Not used in current implementation
	return nil
}

func (m *mockResourceInterface) Get(ctx context.Context, name string, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	key := fmt.Sprintf("%s/%s", m.namespace, name)
	if err, exists := m.errors[key]; exists {
		return nil, err
	}
	if obj, exists := m.created[key]; exists {
		return obj, nil
	}
	return nil, k8serrors.NewNotFound(m.gvr.GroupResource(), name)
}

func (m *mockResourceInterface) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockResourceInterface) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockResourceInterface) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockResourceInterface) Apply(ctx context.Context, name string, obj *unstructured.Unstructured, opts metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	// Not used in current implementation
	return nil, nil
}

func (m *mockResourceInterface) ApplyStatus(ctx context.Context, name string, obj *unstructured.Unstructured, opts metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	// Not used in current implementation
	return nil, nil
}

// Test helper functions
func createTestPipelineRun(name, namespace string, annotations map[string]string, condition *apis.Condition) *pipelinev1.PipelineRun {
	pr := &pipelinev1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
			UID:         types.UID("test-uid"),
		},
		Spec:   pipelinev1.PipelineRunSpec{},
		Status: pipelinev1.PipelineRunStatus{},
	}

	if condition != nil {
		pr.Status.Conditions = []apis.Condition{*condition}
	}

	return pr
}

func createTestAppWrapper(name, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "workload.codeflare.dev/v1beta2",
			"kind":       "AppWrapper",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
}

func createMockReconciler(lister v1.PipelineRunLister, dynamicClient dynamic.Interface) *Reconciler {
	return &Reconciler{
		pipelineRunLister: lister,
		dynamicClient:     dynamicClient,
	}
}

func assertAppWrapperCreated(t *testing.T, client *mockDynamicClient, expectedName string) {
	found := false
	for _, obj := range client.created {
		if obj.GetName() == expectedName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected AppWrapper %s to be created, but it was not found", expectedName)
	}
}

func assertResourceRequests(t *testing.T, aw *unstructured.Unstructured, expected map[string]string) {
	// Debug: Print the full AppWrapper structure
	t.Logf("AppWrapper Object: %+v", aw.Object)

	rawComponents, found, err := unstructured.NestedFieldNoCopy(aw.Object, "spec", "components")
	if err != nil || !found {
		t.Errorf("Failed to get components from AppWrapper")
		return
	}
	var components []interface{}
	switch v := rawComponents.(type) {
	case []interface{}:
		components = v
	case []map[string]interface{}:
		components = make([]interface{}, len(v))
		for i := range v {
			components[i] = v[i]
		}
	default:
		t.Errorf("components is of unexpected type %T", rawComponents)
		return
	}
	if len(components) == 0 {
		t.Errorf("Components is empty")
		return
	}

	component, ok := components[0].(map[string]interface{})
	if !ok {
		t.Errorf("Component is not a map[string]interface{}")
		return
	}
	template, found, err := unstructured.NestedMap(component, "template")
	if err != nil {
		t.Errorf("Error getting template: %v", err)
		return
	}
	if !found {
		t.Errorf("Template not found in component")
		return
	}

	spec, found, err := unstructured.NestedMap(template, "spec")
	if err != nil {
		t.Errorf("Error getting spec: %v", err)
		return
	}
	if !found {
		t.Errorf("Spec not found in template")
		return
	}

	appwrapper, found, err := unstructured.NestedMap(spec, "appwrapper")
	if err != nil {
		t.Errorf("Error getting appwrapper: %v", err)
		return
	}
	if !found {
		t.Errorf("Appwrapper not found in spec")
		return
	}

	containers, found, err := unstructured.NestedSlice(appwrapper, "spec", "containers")
	if err != nil {
		t.Errorf("Error getting containers: %v", err)
		return
	}
	if !found {
		t.Errorf("Containers not found in appwrapper spec")
		return
	}
	if len(containers) == 0 {
		t.Errorf("Containers slice is empty")
		return
	}

	container, ok := containers[0].(map[string]interface{})
	if !ok {
		t.Errorf("Container is not a map[string]interface{}")
		return
	}
	requests, found, err := unstructured.NestedMap(container, "resources", "requests")
	if err != nil {
		t.Errorf("Error getting resource requests: %v", err)
		return
	}
	if !found {
		t.Errorf("Resource requests not found in container")
		return
	}

	for key, expectedValue := range expected {
		if actualValue, exists := requests[key]; !exists {
			t.Errorf("Expected resource request %s not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected resource request %s to be %s, got %v", key, expectedValue, actualValue)
		}
	}
}

// Test cases
func TestReconcileKind_AppWrapperExists(t *testing.T) {
	ctx := context.Background()
	pipelineRun := createTestPipelineRun("test-pr", "test-ns", nil, nil)
	existingAppWrapper := createTestAppWrapper("test-pr-appwrapper", "test-ns")

	mockDynamic := &mockDynamicClient{
		created: map[string]*unstructured.Unstructured{
			"test-ns/test-pr-appwrapper": existingAppWrapper,
		},
		errors: make(map[string]error),
	}
	reconciler := createMockReconciler(&mockPipelineRunLister{}, mockDynamic)

	// Set up config in context
	cfg := &config.Config{DefaultQueueName: "default-queue"}
	ctx = config.ToContext(ctx, cfg)

	event := reconciler.ReconcileKind(ctx, pipelineRun)
	if event != nil {
		t.Errorf("Expected no event when AppWrapper exists, got %v", event)
	}

	// Verify no new AppWrapper was created
	if len(mockDynamic.created) > 1 {
		t.Error("Expected no new AppWrapper to be created when one already exists")
	}
}

func TestReconcileKind_DynamicClientError(t *testing.T) {
	ctx := context.Background()
	pipelineRun := createTestPipelineRun("test-pr", "test-ns", nil, nil)

	mockDynamic := &mockDynamicClient{
		created: make(map[string]*unstructured.Unstructured),
		errors: map[string]error{
			"test-ns/test-pr-appwrapper": fmt.Errorf("dynamic client error"),
		},
	}
	reconciler := createMockReconciler(&mockPipelineRunLister{}, mockDynamic)

	// Set up config in context
	cfg := &config.Config{DefaultQueueName: "default-queue"}
	ctx = config.ToContext(ctx, cfg)

	event := reconciler.ReconcileKind(ctx, pipelineRun)
	if event == nil {
		t.Error("Expected error event when dynamic client fails, got nil")
	}
}

func TestReconcileKind_ResourceAnnotations(t *testing.T) {
	testCases := []struct {
		name        string
		annotations map[string]string
		expected    map[string]string
	}{
		{
			name: "cpu only",
			annotations: map[string]string{
				"kueue.tekton.dev/requests-cpu": "500m",
			},
			expected: map[string]string{
				"cpu": "500m",
			},
		},
		{
			name: "memory only",
			annotations: map[string]string{
				"kueue.tekton.dev/requests-memory": "512Mi",
			},
			expected: map[string]string{
				"memory": "512Mi",
			},
		},
		{
			name: "cpu and memory",
			annotations: map[string]string{
				"kueue.tekton.dev/requests-cpu":    "1000m",
				"kueue.tekton.dev/requests-memory": "1Gi",
			},
			expected: map[string]string{
				"cpu":    "1000m",
				"memory": "1Gi",
			},
		},
		{
			name: "all resources",
			annotations: map[string]string{
				"kueue.tekton.dev/requests-cpu":               "1000m",
				"kueue.tekton.dev/requests-memory":            "1Gi",
				"kueue.tekton.dev/requests-storage":           "10Gi",
				"kueue.tekton.dev/requests-ephemeral-storage": "5Gi",
			},
			expected: map[string]string{
				"cpu":               "1000m",
				"memory":            "1Gi",
				"storage":           "10Gi",
				"ephemeral-storage": "5Gi",
			},
		},
		{
			name:        "no annotations",
			annotations: map[string]string{},
			expected:    map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pipelineRun := createTestPipelineRun("test-pr", "test-ns", tc.annotations, nil)

			mockDynamic := &mockDynamicClient{
				created: make(map[string]*unstructured.Unstructured),
				errors:  make(map[string]error),
			}
			reconciler := createMockReconciler(&mockPipelineRunLister{}, mockDynamic)

			// Set up config in context
			cfg := &config.Config{DefaultQueueName: "default-queue"}
			testCtx := config.ToContext(context.Background(), cfg)

			event := reconciler.ReconcileKind(testCtx, pipelineRun)
			if event == nil {
				t.Error("Expected success event for resource annotations test, got nil")
			}

			createdAppWrapper := mockDynamic.created["test-ns/test-pr-appwrapper"]
			if createdAppWrapper == nil {
				t.Fatal("Expected AppWrapper to be created")
			}

			assertResourceRequests(t, createdAppWrapper, tc.expected)
		})
	}
}

func TestReconcileKind_EndToEndFlow(t *testing.T) {
	ctx := context.Background()
	annotations := map[string]string{
		"kueue.tekton.dev/requests-cpu":    "1000m",
		"kueue.tekton.dev/requests-memory": "1Gi",
	}
	pipelineRun := createTestPipelineRun("test-pr", "test-ns", annotations, nil)

	mockDynamic := &mockDynamicClient{
		created: make(map[string]*unstructured.Unstructured),
		errors:  make(map[string]error),
	}
	reconciler := createMockReconciler(&mockPipelineRunLister{}, mockDynamic)

	// Set up config in context
	cfg := &config.Config{DefaultQueueName: "default-queue"}
	ctx = config.ToContext(ctx, cfg)

	event := reconciler.ReconcileKind(ctx, pipelineRun)
	if event == nil {
		t.Error("Expected success event for end-to-end flow, got nil")
	}

	createdAppWrapper := mockDynamic.created["test-ns/test-pr-appwrapper"]
	if createdAppWrapper == nil {
		t.Fatal("Expected AppWrapper to be created")
	}

	// Verify API version and kind
	if createdAppWrapper.GetAPIVersion() != "workload.codeflare.dev/v1beta2" {
		t.Errorf("Expected API version %s, got %s", "workload.codeflare.dev/v1beta2", createdAppWrapper.GetAPIVersion())
	}
	if createdAppWrapper.GetKind() != "AppWrapper" {
		t.Errorf("Expected kind 'AppWrapper', got %s", createdAppWrapper.GetKind())
	}

	// Verify labels
	labels := createdAppWrapper.GetLabels()
	if labels["kueue.x-k8s.io/queue-name"] != "default-queue" {
		t.Errorf("Expected queue label to be 'default-queue', got %s", labels["kueue.x-k8s.io/queue-name"])
	}

	// Verify owner references
	ownerRefs := createdAppWrapper.GetOwnerReferences()
	if len(ownerRefs) != 1 {
		t.Errorf("Expected 1 owner reference, got %d", len(ownerRefs))
	}

	ownerRef := ownerRefs[0]
	if ownerRef.APIVersion != pipelinev1.SchemeGroupVersion.String() {
		t.Errorf("Expected owner API version %s, got %s", pipelinev1.SchemeGroupVersion.String(), ownerRef.APIVersion)
	}
	if ownerRef.Kind != "PipelineRun" {
		t.Errorf("Expected owner kind 'PipelineRun', got %s", ownerRef.Kind)
	}
	if ownerRef.Name != "test-pr" {
		t.Errorf("Expected owner name 'test-pr', got %s", ownerRef.Name)
	}
	if ownerRef.Controller == nil || !*ownerRef.Controller {
		t.Error("Expected owner reference to be controller")
	}
	if ownerRef.UID != "test-uid" {
		t.Errorf("Expected owner UID 'test-uid', got %s", ownerRef.UID)
	}
}

// Benchmark tests
func BenchmarkReconcileKind_SuccessfulCreation(b *testing.B) {
	annotations := map[string]string{
		"kueue.tekton.dev/requests-cpu":    "1000m",
		"kueue.tekton.dev/requests-memory": "1Gi",
	}
	pipelineRun := createTestPipelineRun("test-pr", "test-ns", annotations, nil)

	mockDynamic := &mockDynamicClient{
		created: make(map[string]*unstructured.Unstructured),
		errors:  make(map[string]error),
	}
	reconciler := createMockReconciler(&mockPipelineRunLister{}, mockDynamic)

	// Set up config in context
	cfg := &config.Config{DefaultQueueName: "default-queue"}
	ctx := config.ToContext(context.Background(), cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset the mock for each iteration
		mockDynamic.created = make(map[string]*unstructured.Unstructured)
		err := reconciler.ReconcileKind(ctx, pipelineRun)
		if err != nil {
			b.Errorf("Expected no error, got %v", err)
		}
	}
}
