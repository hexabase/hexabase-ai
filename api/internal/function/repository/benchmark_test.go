package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/function/domain"
	"github.com/hexabase/hexabase-ai/api/internal/function/repository/fission"
	"github.com/hexabase/hexabase-ai/api/internal/function/repository/knative"
	"github.com/hexabase/hexabase-ai/api/internal/function/repository/mock"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

// BenchmarkProviderComparison compares the performance of different providers
func BenchmarkProviderComparison(b *testing.B) {
	ctx := context.Background()
	
	// Create fake kubernetes clients for Knative
	scheme := runtime.NewScheme()
	kubeClient := fake.NewSimpleClientset()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)
	
	providers := map[string]domain.Provider{
		"mock":    mock.NewFunctionProvider(),
		"fission": fission.NewProvider("http://controller.fission", "default"),
		"knative": knative.NewProvider(kubeClient, dynamicClient, "default"),
	}

	// Test function spec
	spec := &domain.FunctionSpec{
		Name:      "bench-func",
		Namespace: "bench-ns",
		Runtime:   domain.RuntimePython,
		Handler:   "main.handler",
		SourceCode: `def handler(context):
			return {"status": 200, "body": "Hello World"}`,
		Resources: domain.FunctionResourceRequirements{
			Memory: "256Mi",
			CPU:    "100m",
		},
	}

	// Benchmark function creation
	for name, provider := range providers {
		b.Run(fmt.Sprintf("CreateFunction_%s", name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				spec.Name = fmt.Sprintf("bench-func-%d", i)
				_, _ = provider.CreateFunction(ctx, spec)
			}
		})
	}

	// Benchmark function invocation
	req := &domain.InvokeRequest{
		Method: "GET",
		Path:   "/",
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
	}

	for name, provider := range providers {
		b.Run(fmt.Sprintf("InvokeFunction_%s", name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = provider.InvokeFunction(ctx, "bench-func", req)
			}
		})
	}

	// Benchmark cold start simulation
	for name, provider := range providers {
		b.Run(fmt.Sprintf("ColdStart_%s", name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Simulate cold start by creating and immediately invoking
				funcName := fmt.Sprintf("cold-func-%d", i)
				spec.Name = funcName
				
				_, err := provider.CreateFunction(ctx, spec)
				if err == nil {
					// Give the function time to be ready (simulate deployment)
					time.Sleep(100 * time.Millisecond)
					
					start := time.Now()
					_, _ = provider.InvokeFunction(ctx, funcName, req)
					elapsed := time.Since(start)
					
					b.ReportMetric(float64(elapsed.Milliseconds()), "ms/cold-start")
				}
			}
		})
	}

	// Benchmark concurrent invocations
	for name, provider := range providers {
		b.Run(fmt.Sprintf("ConcurrentInvoke_%s", name), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, _ = provider.InvokeFunction(ctx, "bench-func", req)
				}
			})
		})
	}
}

// BenchmarkProviderCapabilities tests the performance of capability checks
func BenchmarkProviderCapabilities(b *testing.B) {
	providers := map[string]domain.Provider{
		"mock":    mock.NewFunctionProvider(),
		"fission": fission.NewProvider("http://controller.fission", "default"),
		"knative": knative.NewProvider(fake.NewSimpleClientset(), dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()), "default"),
	}

	for name, provider := range providers {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = provider.GetCapabilities()
			}
		})
	}
}

// TestProviderColdStartComparison compares cold start times
func TestProviderColdStartComparison(t *testing.T) {
	t.Skip("Skipping cold start comparison in unit tests - run benchmarks instead")
	
	// This test would be run in integration tests with real providers
	// It measures actual cold start times and compares them
	
	providers := map[string]struct {
		provider    domain.Provider
		expectedMs  int
		toleranceMs int
	}{
		"fission": {
			provider:    fission.NewProvider("http://controller.fission", "default"),
			expectedMs:  100,  // Fission typically 50-200ms
			toleranceMs: 150,
		},
		"knative": {
			provider:    knative.NewProvider(nil, nil, "default"),
			expectedMs:  2000, // Knative typically 2-5s
			toleranceMs: 3000,
		},
	}

	for name, testCase := range providers {
		t.Run(name, func(t *testing.T) {
			// Would measure actual cold start time and compare
			// against expected values with tolerance
			_ = testCase // TODO: Implement actual cold start measurement
		})
	}
}

// TestProviderResourceEfficiency tests resource usage
func TestProviderResourceEfficiency(t *testing.T) {
	// This would test memory and CPU usage of functions
	// deployed on different providers
	
	providers := map[string]struct {
		provider         domain.Provider
		expectedMemoryMB int
		expectedCPUCores float64
	}{
		"fission": {
			provider:         fission.NewProvider("http://controller.fission", "default"),
			expectedMemoryMB: 128, // Fission is more lightweight
			expectedCPUCores: 0.1,
		},
		"knative": {
			provider:         knative.NewProvider(nil, nil, "default"),
			expectedMemoryMB: 256, // Knative has more overhead
			expectedCPUCores: 0.25,
		},
	}

	for name, tc := range providers {
		t.Run(name, func(t *testing.T) {
			// Would deploy a function and measure actual resource usage
			// This is a placeholder for integration tests
			t.Logf("Provider %s expects %dMB memory and %.2f CPU cores", 
				name, tc.expectedMemoryMB, tc.expectedCPUCores)
		})
	}
}