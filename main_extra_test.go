package main

import (
	"reflect"
	"testing"
)

func TestToEnvs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want map[string]interface{}
	}{
		{
			name: "normal key value",
			in:   []string{"K=V"},
			want: map[string]interface{}{"K": "V"},
		},
		{
			name: "value contains equals sign",
			in:   []string{"K=a=b"},
			want: map[string]interface{}{"K": "a=b"},
		},
		{
			name: "missing separator is skipped",
			in:   []string{"FOO"},
			want: map[string]interface{}{},
		},
		{
			name: "mixed valid and invalid",
			in:   []string{"A=1", "BAD", "B=x=y"},
			want: map[string]interface{}{"A": "1", "B": "x=y"},
		},
		{
			name: "empty input",
			in:   []string{},
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toEnvs(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toEnvs(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestSplitKeyValuePairs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		sep  string
		want map[string]interface{}
	}{
		{
			name: "normal key value",
			in:   []string{"K=V"},
			sep:  "=",
			want: map[string]interface{}{"K": "V"},
		},
		{
			name: "value contains separator",
			in:   []string{"K=a=b"},
			sep:  "=",
			want: map[string]interface{}{"K": "a=b"},
		},
		{
			name: "missing separator is skipped",
			in:   []string{"FOO"},
			sep:  "=",
			want: map[string]interface{}{},
		},
		{
			name: "empty input",
			in:   []string{},
			sep:  "=",
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitKeyValuePairs(tt.in, tt.sep)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitKeyValuePairs(%v, %q) = %v, want %v", tt.in, tt.sep, got, tt.want)
			}
		})
	}
}
