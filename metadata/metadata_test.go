package metadata

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mackerelio/mackerel-agent/config"
)

func TestMetadataGeneratorFetch(t *testing.T) {
	tests := []struct {
		command  string
		metadata string
		err      bool
	}{
		{
			command:  `go run testdata/json.go -exit-code 0 -metadata "{}"`,
			metadata: `{}`,
			err:      false,
		},
		{
			command:  `go run testdata/json.go -exit-code 1 -metadata "{}"`,
			metadata: ``,
			err:      true,
		},
		{
			command:  `go run testdata/json.go -exit-code 0 -metadata '{"example": "metadata", "foo": [100, 200, {}, null]}'`,
			metadata: `{"example":"metadata","foo":[100,200,{},null]}`,
			err:      false,
		},
		{
			command:  `go run testdata/json.go -exit-code 0 -metadata '{"example": metadata"}'`,
			metadata: ``,
			err:      true,
		},
		{
			command:  `go run testdata/json.go -exit-code 0 -metadata '"foobar"'`,
			metadata: `"foobar"`,
			err:      false,
		},
		{
			command:  `go run testdata/json.go -exit-code 0 -metadata foobar`,
			metadata: ``,
			err:      true,
		},
		{
			command:  `go run testdata/json.go -exit-code 0 -metadata 262144`,
			metadata: `262144`,
			err:      false,
		},
		{
			command:  `go run testdata/json.go -exit-code 0 -metadata true`,
			metadata: `true`,
			err:      false,
		},
		{
			command:  `go run testdata/json.go -exit-code 0 -metadata null`,
			metadata: `null`,
			err:      false,
		},
	}
	for _, test := range tests {
		g := Generator{
			Config: &config.MetadataPlugin{
				Command: test.command,
			},
		}
		metadata, err := g.Fetch()
		if err != nil {
			if !test.err {
				t.Errorf("error occurred unexpectedly on command %q", test.command)
			}
		} else {
			if test.err {
				t.Errorf("error did not occurr but error expected on command %q", test.command)
			}
			metadataStr, _ := json.Marshal(metadata)
			if string(metadataStr) != test.metadata {
				t.Errorf("metadata should be %q but got %q", test.metadata, string(metadataStr))
			}
		}
	}
}

func TestMetadataGeneratorSaveDiffers(t *testing.T) {
	tests := []struct {
		prevmetadata string
		metadata     string
		differs      bool
	}{
		{
			prevmetadata: `{}`,
			metadata:     `{}`,
			differs:      false,
		},
		{
			prevmetadata: `{ "foo": [ 100, 200, null, {} ] }`,
			metadata:     `{"foo":[100,200,null,{}]}`,
			differs:      false,
		},
		{
			prevmetadata: `null`,
			metadata:     `{}`,
			differs:      true,
		},
		{
			prevmetadata: `[]`,
			metadata:     `{}`,
			differs:      true,
		},
	}
	for _, test := range tests {
		g := Generator{}
		var prevmetadata interface{}
		_ = json.Unmarshal([]byte(test.prevmetadata), &prevmetadata)

		if err := g.Save(prevmetadata); err != nil {
			t.Errorf("Error should not occur in Save() but got: %s", err.Error())
		}

		var metadata interface{}
		_ = json.Unmarshal([]byte(test.metadata), &metadata)

		got := g.Differs(metadata)
		if got != test.differs {
			t.Errorf("Differs() should return %t but got %t for %v, %v", test.differs, got, prevmetadata, metadata)
		}
	}
}

func TestMetadataGeneratorInterval(t *testing.T) {
	tests := []struct {
		interval *int32
		expected time.Duration
	}{
		{
			interval: pint(0),
			expected: 1 * time.Minute,
		},
		{
			interval: pint(1),
			expected: 1 * time.Minute,
		},
		{
			interval: pint(30),
			expected: 30 * time.Minute,
		},
		{
			interval: nil,
			expected: 10 * time.Minute,
		},
	}
	for _, test := range tests {
		g := Generator{
			Config: &config.MetadataPlugin{
				ExecutionInterval: test.interval,
			},
		}
		if g.Interval() != test.expected {
			t.Errorf("interval should be %v but got: %v", test.expected, g.Interval())
		}
	}
}

func pint(i int32) *int32 {
	return &i
}
