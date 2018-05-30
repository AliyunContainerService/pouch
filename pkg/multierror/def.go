package multierror

import (
	"fmt"
	"strings"
)

// Multierrors contains a slice of errors.
type Multierrors struct {
	errs []error
}

// Append adds the errors into the list.
func (m *Multierrors) Append(errs ...error) {
	m.errs = append(m.errs, errs...)
}

// Size returns the count of list of errors.
func (m *Multierrors) Size() int {
	return len(m.errs)
}

// Error returns the combined error messages.
func (m *Multierrors) Error() string {
	if len(m.errs) == 0 {
		return fmt.Sprintf("no error")
	}

	if len(m.errs) == 1 {
		return fmt.Sprintf("%s", m.errs[0])
	}

	serrs := make([]string, len(m.errs))
	for i, err := range m.errs {
		serrs[i] = fmt.Sprintf("* %s", err)
	}
	return fmt.Sprintf("%d errors:\n\n%s", len(m.errs), strings.Join(serrs, "\n"))
}
