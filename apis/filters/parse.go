package filters

import (
	"encoding/json"
	"errors"
	"strings"
)

// Args stores filter arguments as map key:{map key: bool}.
// It contains an aggregation of the map of arguments (which are in the form
// of -f 'key=value') based on the key, and stores values for the same key
// in a map with string keys and boolean values.
// e.g given -f 'label=label1=1' -f 'label=label2=2' -f 'image.name=ubuntu'
// the args will be {"image.name":{"ubuntu":true},"label":{"label1=1":true,"label2=2":true}}
type Args struct {
	fields map[string]map[string]bool
}

// KeyValuePair is used to initialize a new Args
type KeyValuePair struct {
	Key   string
	Value string
}

// Arg creates a new KeyValuePair for initializing Args
func Arg(key, value string) KeyValuePair {
	return KeyValuePair{Key: key, Value: value}
}

// NewArgs returns a new Args populated with the initial args
func NewArgs(initialArgs ...KeyValuePair) Args {
	args := Args{fields: map[string]map[string]bool{}}
	for _, arg := range initialArgs {
		args.Add(arg.Key, arg.Value)
	}
	return args
}

// Get returns the list of values associated with the key
func (args Args) Get(key string) []string {
	values := args.fields[key]
	if values == nil {
		return make([]string, 0)
	}
	slice := make([]string, 0, len(values))
	for key := range values {
		slice = append(slice, key)
	}
	return slice
}

// Add a new value to the set of values
func (args Args) Add(key, value string) {
	if _, ok := args.fields[key]; ok {
		args.fields[key][value] = true
	} else {
		args.fields[key] = map[string]bool{value: true}
	}
}

// Del removes a value from the set
func (args Args) Del(key, value string) {
	if _, ok := args.fields[key]; ok {
		delete(args.fields[key], value)
		if len(args.fields[key]) == 0 {
			delete(args.fields, key)
		}
	}
}

// Len returns the number of fields in the arguments.
func (args Args) Len() int {
	return len(args.fields)
}

// ExactMatch returns true if the source matches exactly one of the filters.
func (args Args) ExactMatch(field, source string) bool {
	fieldValues, ok := args.fields[field]
	//do not filter if there is no filter set or cannot determine filter
	if !ok || len(fieldValues) == 0 {
		return true
	}

	// try to match full name value to avoid O(N) regular expression matching
	return fieldValues[source]
}

// MarshalJSON returns a JSON byte representation of the Args
func (args Args) MarshalJSON() ([]byte, error) {
	if len(args.fields) == 0 {
		return []byte{}, nil
	}
	return json.Marshal(args.fields)
}

// UnmarshalJSON populates the Args from JSON encode bytes
func (args Args) UnmarshalJSON(raw []byte) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, &args.fields)
}

// ErrBadFormat is an error returned when a filter is not in the form key=value
var ErrBadFormat = errors.New("bad format of filter (expected name=value)")

// ParseFlag parses a key=value string and adds it to an Args.
func ParseFlag(arg string, prev Args) (Args, error) {
	filters := prev
	if len(arg) == 0 {
		return filters, nil
	}

	if !strings.Contains(arg, "=") {
		return filters, ErrBadFormat
	}

	f := strings.SplitN(arg, "=", 2)

	name := strings.ToLower(strings.TrimSpace(f[0]))
	value := strings.TrimSpace(f[1])

	filters.Add(name, value)

	return filters, nil
}

// ToParam packs the Args into a string for easy transport from client to server.
func ToParam(a Args) (string, error) {
	if a.Len() == 0 {
		return "", nil
	}

	buf, err := json.Marshal(a)
	return string(buf), err
}

// FromParam decodes a JSON encoded string into Args
func FromParam(p string) (Args, error) {
	args := NewArgs()

	if p == "" {
		return args, nil
	}

	raw := []byte(p)
	err := json.Unmarshal(raw, &args)
	if err != nil {
		return args, err
	}
	return args, nil
}
