package server

import (
	"fmt"
)

func validationName(name string) error {
	if name == "" {
		return fmt.Errorf("name is empty string")
	}

	// TODO add more validations
	return nil
}
