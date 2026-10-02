package core

import (
	"reflect"
	"testing"
)

func TestJSONMapScan(t *testing.T) {
	tests := []struct {
		name   string
		source interface{}
		want   JSONMap
	}{
		{
			name:   "bytes",
			source: []byte(`{"theme":"dark","enabled":true}`),
			want:   JSONMap{"theme": "dark", "enabled": true},
		},
		{
			name:   "string",
			source: `{"count":3}`,
			want:   JSONMap{"count": float64(3)},
		},
		{
			name:   "null",
			source: nil,
			want:   JSONMap{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got JSONMap
			if err := got.Scan(test.source); err != nil {
				t.Fatalf("Scan() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Scan() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestJSONMapValue(t *testing.T) {
	value, err := (JSONMap{"theme": "dark"}).Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}
	if value != `{"theme":"dark"}` {
		t.Errorf("Value() = %s, want JSON object", value)
	}
}
