package lib

import "testing"

func TestNewRequest(t *testing.T) {
	var tests = []struct {
		input string
		valid bool
	}{
		{"", false},
		{"{}", false},
		{`{"still": "flying"}`, false},
		{`{"command": "ban"}`, false},
		{`{"command": "ban", "host": "example.org"}`, false},
		{`{"command": "ban", "host": "example.org", "value": ""}`, false},
		{`{"command": "ban", "host": "example.org", "value": []}`, false},
		{`{"command": "ban", "host": "example.org", "value": ["expr"]}`, true},
		{`{"command": "ban", "host": "example.org", "value": ["expr", "expr2"]}`, true},
		{`{"command": "ban"}`, false},
		{`{"command": "ban.url", "host": "example.org"}`, false},
		{`{"command": "ban.url", "host": "example.org", "value": ""}`, false},
		{`{"command": "ban.url", "host": "example.org", "value": []}`, false},
		{`{"command": "ban.url", "host": "example.org", "value": ["url"]}`, true},
		{`{"command": "ban.url", "host": "example.org", "value": ["url1", "url2"]}`, true},
		{`{"command": "purge"}`, false},
		{`{"command": "purge", "host": "example.org"}`, false},
		{`{"command": "purge", "host": "example.org", "value": ""}`, false},
		{`{"command": "purge", "host": "example.org", "value": []}`, false},
		{`{"command": "purge", "host": "example.org", "value": ["path"]}`, true},
		{`{"command": "purge", "host": "example.org", "value": ["path1", "path2"]}`, true},
		{`{"command": "xkey"}`, false},
		{`{"command": "xkey", "host": "example.org"}`, false},
		{`{"command": "xkey", "host": "example.org", "value": ""}`, false},
		{`{"command": "xkey", "host": "example.org", "value": []}`, false},
		{`{"command": "xkey", "host": "example.org", "value": ["key"]}`, true},
		{`{"command": "xkey", "host": "example.org", "value": ["key1", "key2"]}`, true},
		{`{"command": "xkey.soft"}`, false},
		{`{"command": "xkey.soft", "host": "example.org"}`, false},
		{`{"command": "xkey.soft", "host": "example.org", "value": ""}`, false},
		{`{"command": "xkey.soft", "host": "example.org", "value": []}`, false},
		{`{"command": "xkey.soft", "host": "example.org", "value": ["key"]}`, true},
		{`{"command": "xkey.soft", "host": "example.org", "value": ["key1", "key2"]}`, true},
	}

	for _, test := range tests {
		_, err := NewRequest(test.input)

		if test.valid && err != nil {
			t.Errorf("Input \"%s\" should be recognized as valid, but isn't.", test.input)
		}

		if !test.valid && err == nil {
			t.Errorf("Input \"%s\" should be recognized as invalid, but isn't.", test.input)
		}
	}

}
