package shared

import (
	"testing"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
)

func TestNewSharedContext(t *testing.T) {
	mockProvider := &vmock.MockCloud{}
	
	ctx := NewSharedContext(mockProvider)
	
	if ctx == nil {
		t.Fatal("NewSharedContext should not return nil")
	}
	
	// Verify provider is set
	provider := ctx.Provider()
	if provider == nil {
		t.Fatal("Provider should not be nil after NewSharedContext")
	}
	
	if provider != mockProvider {
		t.Error("Provider returned does not match the one passed to NewSharedContext")
	}
}

func TestSharedContext_Provider(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *SharedContext
		wantNil     bool
		expectPanic bool
	}{
		{
			name: "provider exists",
			setup: func() *SharedContext {
				mockProvider := &vmock.MockCloud{}
				return NewSharedContext(mockProvider)
			},
			wantNil: false,
		},
		{
			name: "provider not set",
			setup: func() *SharedContext {
				return &SharedContext{
					Context: base.Context{},
				}
			},
			wantNil: true,
		},
		{
			name: "provider set to nil",
			setup: func() *SharedContext {
				ctx := &SharedContext{}
				ctx.SetKV(Provider, nil)
				return ctx
			},
			wantNil:     true,
			expectPanic: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic but did not occur")
					}
				}()
				_ = ctx.Provider()
				return
			}
			
			provider := ctx.Provider()
			
			if tt.wantNil && provider != nil {
				t.Errorf("Provider() should return nil, got %v", provider)
			}
			if !tt.wantNil && provider == nil {
				t.Error("Provider() should not return nil")
			}
		})
	}
}

func TestSharedContext_SetAndGetProvider(t *testing.T) {
	mockProvider1 := &vmock.MockCloud{}
	mockProvider2 := &vmock.MockCloud{}
	
	ctx := NewSharedContext(mockProvider1)
	
	// Verify initial provider
	if ctx.Provider() != mockProvider1 {
		t.Error("Initial provider mismatch")
	}
	
	// Update provider
	ctx.SetKV(Provider, mockProvider2)
	
	// Verify updated provider
	if ctx.Provider() != mockProvider2 {
		t.Error("Updated provider mismatch")
	}
}

func TestSharedContext_InheritFromBaseContext(t *testing.T) {
	mockProvider := &vmock.MockCloud{}
	ctx := NewSharedContext(mockProvider)
	
	// Test that SharedContext inherits base.Context methods
	testKey := "testKey"
	testValue := "testValue"
	
	// Test SetKV
	ctx.SetKV(testKey, testValue)
	
	// Test Value
	val, ok := ctx.Value(testKey)
	if !ok {
		t.Error("Value() should return true for existing key")
	}
	if val != testValue {
		t.Errorf("Value() = %v, want %v", val, testValue)
	}
	
	// Test Range
	count := 0
	ctx.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	
	// Should have at least Provider and testKey
	if count < 2 {
		t.Errorf("Range() should iterate over at least 2 entries, got %d", count)
	}
}
