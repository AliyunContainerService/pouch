package filters

import (
	"testing"
)

func TestParseArgs(t *testing.T) {
	// equivalent of `pouch ps -f 'created=today' -f 'image.name=ubuntu*' -f 'image.name=*untu'`
	flagArgs := []string{
		"created=today",
		"image.name=ubuntu*",
		"image.name=*untu",
	}
	var (
		args = NewArgs()
		err  error
	)

	for i := range flagArgs {
		args, err = ParseFlag(flagArgs[i], args)
		if err != nil {
			t.Fatalf("ParseFlag got err: %v", err)
		}
	}

	if len(args.Get("created")) != 1 {
		t.Fatalf("got unexpected created keys: %v", args.Get("created"))
	}
	if len(args.Get("image.name")) != 2 {
		t.Fatalf("got unexpected image.name keys: %v", args.Get("image.name"))
	}
}

func TestAdd(t *testing.T) {
	f := NewArgs()
	f.Add("status", "running")
	v := f.fields["status"]
	if len(v) != 1 || !v["running"] {
		t.Fatalf("Expected to include a running status, got %v", v)
	}

	f.Add("status", "paused")
	if len(v) != 2 || !v["paused"] {
		t.Fatalf("Expected to include a paused status, got %v", v)
	}
}

func TestDel(t *testing.T) {
	f := NewArgs()
	f.Add("status", "running")
	f.Del("status", "running")
	v := f.fields["status"]
	if v["running"] {
		t.Fatal("Expected to not include a running status filter, got true")
	}
}

func TestLen(t *testing.T) {
	f := NewArgs()
	if f.Len() != 0 {
		t.Fatal("Expected to not include any field")
	}
	f.Add("status", "running")
	if f.Len() != 1 {
		t.Fatal("Expected to include one field")
	}
}

func TestExactMatch(t *testing.T) {
	f := NewArgs()

	if !f.ExactMatch("status", "running") {
		t.Fatal("Expected to match `running` when there are no filters, got false")
	}

	f.Add("status", "running")
	f.Add("status", "pause*")

	if !f.ExactMatch("status", "running") {
		t.Fatal("Expected to match `running` with one of the filters, got false")
	}

	if f.ExactMatch("status", "paused") {
		t.Fatal("Expected to not match `paused` with one of the filters, got true")
	}
}

func TestToParam(t *testing.T) {
	fields := map[string]map[string]bool{
		"created":    {"today": true},
		"image.name": {"ubuntu*": true, "*untu": true},
	}
	a := Args{fields: fields}

	_, err := ToParam(a)
	if err != nil {
		t.Errorf("failed to marshal the filters: %s", err)
	}
}

func TestFromParam(t *testing.T) {
	invalids := []string{
		"anything",
		"['a','list']",
		"{'key': 'value'}",
		`{"key": "value"}`,
		`{"key": ["value"]}`,
	}
	valid := map[*Args][]string{
		{fields: map[string]map[string]bool{"key": {"value": true}}}: {
			`{"key": {"value": true}}`,
		},
		{fields: map[string]map[string]bool{"key": {"value1": true, "value2": true}}}: {
			`{"key": {"value1": true, "value2": true}}`,
		},
		{fields: map[string]map[string]bool{"key1": {"value1": true}, "key2": {"value2": true}}}: {
			`{"key1": {"value1": true}, "key2": {"value2": true}}`,
		},
	}

	for _, invalid := range invalids {
		if _, err := FromParam(invalid); err == nil {
			t.Fatalf("Expected an error with %v, got nothing", invalid)
		}
	}

	for expectedArgs, matchers := range valid {
		for _, json := range matchers {
			args, err := FromParam(json)
			if err != nil {
				t.Fatal(err)
			}
			if args.Len() != expectedArgs.Len() {
				t.Fatalf("Expected %v, go %v", expectedArgs, args)
			}
			for key, expectedValues := range expectedArgs.fields {
				values := args.Get(key)

				if len(values) != len(expectedValues) {
					t.Fatalf("Expected %v, go %v", expectedArgs, args)
				}

				for _, v := range values {
					if !expectedValues[v] {
						t.Fatalf("Expected %v, go %v", expectedArgs, args)
					}
				}
			}
		}
	}
}
