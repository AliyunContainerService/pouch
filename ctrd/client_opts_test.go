package ctrd

import "testing"

func TestWithInsecureRegistries(t *testing.T) {
	testCases := []struct {
		endpoints []string
		hasError  bool
	}{
		{
			endpoints: []string{"localhost:5000"},
			hasError:  false,
		},
		{
			endpoints: []string{"localhost:5000/v1"},
			hasError:  true,
		},
		{
			endpoints: []string{"myregistry.com"},
			hasError:  false,
		},
		{
			endpoints: []string{"myregistry.com:5000"},
			hasError:  false,
		},
		{
			endpoints: []string{"myregistry.com:5000/v1"},
			hasError:  true,
		},
		{
			endpoints: []string{"http://myregistry.com:5000"},
			hasError:  true,
		},
		{
			endpoints: []string{"dummy://myregistry.com:5000"},
			hasError:  true,
		},
		{
			endpoints: []string{"myregistry.com:65536"},
			hasError:  true,
		},
	}

	for _, tc := range testCases {
		err := WithInsecureRegistries(tc.endpoints)(&clientOpts{})
		if (err != nil) != tc.hasError {
			t.Fatalf("expected hasError = %v, but got error = %v", tc.hasError, err)
		}
	}
}
