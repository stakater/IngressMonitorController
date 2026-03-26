package main

import (
	"os"
	"testing"
)

func TestBuildDefaultNamespaces(t *testing.T) {
	tests := []struct {
		name           string
		watchNamespace string
		wantNil        bool
		wantKeys       []string
	}{
		{
			name:           "empty string returns nil (cluster-scoped cache)",
			watchNamespace: "",
			wantNil:        true,
		},
		{
			name:           "single namespace",
			watchNamespace: "my-namespace",
			wantKeys:       []string{"my-namespace"},
		},
		{
			name:           "multiple namespaces",
			watchNamespace: "ns1,ns2,ns3",
			wantKeys:       []string{"ns1", "ns2", "ns3"},
		},
		{
			name:           "namespaces with surrounding spaces are trimmed",
			watchNamespace: "ns1, ns2 , ns3",
			wantKeys:       []string{"ns1", "ns2", "ns3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDefaultNamespaces(tt.watchNamespace)

			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected non-nil map, got nil")
			}

			if len(got) != len(tt.wantKeys) {
				t.Errorf("expected %d namespaces, got %d: %v", len(tt.wantKeys), len(got), got)
			}

			for _, key := range tt.wantKeys {
				if _, ok := got[key]; !ok {
					t.Errorf("expected key %q in map, got %v", key, got)
				}
			}
		})
	}
}

func TestGetWatchNamespace(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		envSet  bool
		wantNS  string
		wantErr bool
	}{
		{
			name:    "env var not set returns error",
			envSet:  false,
			wantErr: true,
		},
		{
			name:   "empty env var returns empty string (cluster-scoped)",
			envSet: true,
			envVal: "",
			wantNS: "",
		},
		{
			name:   "single namespace",
			envSet: true,
			envVal: "my-namespace",
			wantNS: "my-namespace",
		},
		{
			name:   "multiple namespaces passed through as-is",
			envSet: true,
			envVal: "ns1,ns2",
			wantNS: "ns1,ns2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("WATCH_NAMESPACE")
			if tt.envSet {
				os.Setenv("WATCH_NAMESPACE", tt.envVal)
				defer os.Unsetenv("WATCH_NAMESPACE")
			}

			got, err := getWatchNamespace()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.wantNS {
				t.Errorf("expected %q, got %q", tt.wantNS, got)
			}
		})
	}
}
