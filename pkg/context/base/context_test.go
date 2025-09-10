package base

import (
	"sync"
	"testing"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("NewContext should not return nil")
	}
	if ctx.Ctx == nil {
		t.Fatal("NewContext should initialize Ctx field")
	}
}

func TestContext_SetKV(t *testing.T) {
	tests := []struct {
		name  string
		ctx   *Context
		key   string
		value interface{}
	}{
		{
			name:  "set value on initialized context",
			ctx:   &Context{Ctx: &sync.Map{}},
			key:   "testKey",
			value: "testValue",
		},
		{
			name:  "set value on nil context",
			ctx:   &Context{},
			key:   "testKey",
			value: "testValue",
		},
		{
			name:  "set nil value",
			ctx:   &Context{Ctx: &sync.Map{}},
			key:   "testKey",
			value: nil,
		},
		{
			name:  "set complex value",
			ctx:   &Context{Ctx: &sync.Map{}},
			key:   "complexKey",
			value: struct{ Name string }{Name: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ctx.SetKV(tt.key, tt.value)
			
			if tt.ctx.Ctx == nil {
				t.Fatal("SetKV should initialize Ctx if nil")
			}
			
			val, ok := tt.ctx.Ctx.Load(tt.key)
			if !ok {
				t.Errorf("key %s not found after SetKV", tt.key)
			}
			if val != tt.value {
				t.Errorf("value mismatch: got %v, want %v", val, tt.value)
			}
		})
	}
}

func TestContext_Value(t *testing.T) {
	tests := []struct {
		name      string
		ctx       *Context
		key       string
		wantValue interface{}
		wantOk    bool
	}{
		{
			name: "get existing value",
			ctx: func() *Context {
				c := &Context{Ctx: &sync.Map{}}
				c.Ctx.Store("key1", "value1")
				return c
			}(),
			key:       "key1",
			wantValue: "value1",
			wantOk:    true,
		},
		{
			name:      "get non-existing value",
			ctx:       &Context{Ctx: &sync.Map{}},
			key:       "nonExistingKey",
			wantValue: nil,
			wantOk:    false,
		},
		{
			name:      "get from nil context",
			ctx:       &Context{},
			key:       "anyKey",
			wantValue: nil,
			wantOk:    false,
		},
		{
			name: "get nil value",
			ctx: func() *Context {
				c := &Context{Ctx: &sync.Map{}}
				c.Ctx.Store("nilKey", nil)
				return c
			}(),
			key:       "nilKey",
			wantValue: nil,
			wantOk:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := tt.ctx.Value(tt.key)
			if gotOk != tt.wantOk {
				t.Errorf("Value() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotValue != tt.wantValue {
				t.Errorf("Value() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestContext_Range(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *Context
		setup    func(*Context)
		wantKeys []string
	}{
		{
			name: "range over multiple entries",
			ctx:  &Context{Ctx: &sync.Map{}},
			setup: func(c *Context) {
				c.SetKV("key1", "value1")
				c.SetKV("key2", "value2")
				c.SetKV("key3", "value3")
			},
			wantKeys: []string{"key1", "key2", "key3"},
		},
		{
			name:     "range over empty context",
			ctx:      &Context{Ctx: &sync.Map{}},
			setup:    func(c *Context) {},
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.ctx)
			
			var keys []string
			tt.ctx.Range(func(key, value interface{}) bool {
				keys = append(keys, key.(string))
				return true
			})
			
			if len(keys) != len(tt.wantKeys) {
				t.Errorf("Range() got %d keys, want %d keys", len(keys), len(tt.wantKeys))
				return
			}
			
			// Convert to map for easier comparison
			keyMap := make(map[string]bool)
			for _, k := range keys {
				keyMap[k] = true
			}
			
			for _, wantKey := range tt.wantKeys {
				if !keyMap[wantKey] {
					t.Errorf("Range() missing key %s", wantKey)
				}
			}
		})
	}
}

func TestContext_RangeEarlyExit(t *testing.T) {
	ctx := &Context{Ctx: &sync.Map{}}
	ctx.SetKV("key1", "value1")
	ctx.SetKV("key2", "value2")
	ctx.SetKV("key3", "value3")
	
	count := 0
	ctx.Range(func(key, value interface{}) bool {
		count++
		return count < 2 // Stop after 2 iterations
	})
	
	if count != 2 {
		t.Errorf("Range() should stop early, got %d iterations, want 2", count)
	}
}
