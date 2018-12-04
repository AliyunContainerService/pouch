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

func TestFromFilterOpts(t *testing.T) {
	filterOpts := []string{
		"reference=img1",
		"since=img2",
		"before=img3",
		"reference=img3",
	}

	args, err := FromFilterOpts(filterOpts)
	if err != nil {
		t.Fatal(err)
	}

	images := args.Get("reference")
	if len(images) != 2 {
		t.Fatal("Expected two values of reference key, but got one.")
	}

	if !args.Contains("since") {
		t.Fatal("Excepted get since key, but got none.")
	}

	if !args.Contains("before") {
		t.Fatal("Excepted get before key, but got none.")
	}
}

func TestArgsMatchKVList(t *testing.T) {
	// Not empty sources
	sources := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	matches := map[*Args]string{
		{}: "field",
		{map[string]map[string]bool{
			"created": {"today": true},
			"labels":  {"key1": true}},
		}: "labels",
		{map[string]map[string]bool{
			"created": {"today": true},
			"labels":  {"key1=value1": true}},
		}: "labels",
		{map[string]map[string]bool{
			"created": {"today": true},
			"labels":  {"key1!=value2": true}},
		}: "labels",
	}

	for args, field := range matches {
		if !args.MatchKVList(field, sources) {
			t.Fatalf("Expected true for %v on %v, got false", sources, args)
		}
	}

	differs := map[*Args]string{
		{map[string]map[string]bool{
			"created": {"today": true}},
		}: "created",
		{map[string]map[string]bool{
			"created": {"today": true},
			"labels":  {"key4": true}},
		}: "labels",
		{map[string]map[string]bool{
			"created": {"today": true},
			"labels":  {"key1=value3": true}},
		}: "labels",
		{map[string]map[string]bool{
			"created": {"today": true},
			"labels":  {"key1!=value1": true}},
		}: "labels",
	}

	for args, field := range differs {
		if args.MatchKVList(field, sources) {
			t.Fatalf("Expected false for %v on %v, got true", sources, args)
		}
	}
}

func TestArgsValidate(t *testing.T) {
	tests := []struct {
		name     string
		testArgs Args
		accepted map[string]bool
		wantErr  bool
	}{
		{
			name: "mapping keys are in the accepted set",
			testArgs: Args{
				map[string]map[string]bool{
					"created": {
						"today": true,
					},
				},
			},
			accepted: map[string]bool{
				"created": true,
			},
			wantErr: false,
		},
		{
			name: "mapping keys are not in the accepted set",
			testArgs: Args{
				map[string]map[string]bool{
					"created": {
						"today": true,
					},
				},
			},
			accepted: map[string]bool{
				"created": false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testArgs.Validate(tt.accepted)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestArgsMatch(t *testing.T) {
	source := "today"

	matches := map[*Args]string{
		{}: "field",
		{map[string]map[string]bool{
			"created": {"today": true}},
		}: "today",
		{map[string]map[string]bool{
			"created": {"to*": true}},
		}: "created",
		{map[string]map[string]bool{
			"created": {"to(.*)": true}},
		}: "created",
		{map[string]map[string]bool{
			"created": {"tod": true}},
		}: "created",
		{map[string]map[string]bool{
			"created": {"anything": true, "to*": true}},
		}: "created",
	}

	for args, field := range matches {
		if !args.Match(field, source) {
			t.Fatalf("Expected field %s to match %s", field, source)
		}
	}

	differs := map[*Args]string{
		{map[string]map[string]bool{
			"created": {"tomorrow": true}},
		}: "created",
		{map[string]map[string]bool{
			"created": {"to(day": true}},
		}: "created",
		{map[string]map[string]bool{
			"created": {"tom(.*)": true}},
		}: "created",
		{map[string]map[string]bool{
			"created": {"tom": true}},
		}: "created",
		{map[string]map[string]bool{
			"created": {"today1": true},
			"labels":  {"today": true}},
		}: "created",
	}

	for args, field := range differs {
		if args.Match(field, source) {
			t.Fatalf("Expected field %s to not match %s", field, source)
		}
	}
}
