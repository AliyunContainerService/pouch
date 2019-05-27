package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// reads a file of line terminated key=value pairs, and overrides any keys
// present in the file with additional pairs specified in the override parameter
func readKVStrings(files []string, override []string) ([]string, error) {
	envVariables := []string{}
	for _, ef := range files {
		parsedVars, err := ParseEnvFile(ef)
		if err != nil {
			return nil, err
		}
		envVariables = append(envVariables, parsedVars...)
	}
	// parse the '-e' and '--env' after, to allow override
	envVariables = append(envVariables, override...)

	return envVariables, nil
}

// ParseEnvFile reads a file with environment variables enumerated by lines
func ParseEnvFile(filename string) ([]string, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return []string{}, err
	}
	defer fh.Close()

	lines := []string{}
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		// trim the line from all leading whitespace first
		line := strings.TrimLeft(scanner.Text(), whiteSpaces)
		// line is not empty, and not starting with '#'
		if len(line) > 0 && !strings.HasPrefix(line, "#") {
			data := strings.SplitN(line, "=", 2)

			// trim the front of a variable, but nothing else
			variable := strings.TrimLeft(data[0], whiteSpaces)
			if strings.ContainsAny(variable, whiteSpaces) {
				return []string{}, ErrBadEnvVariable{fmt.Sprintf("variable '%s' has white spaces", variable)}
			}

			if len(data) > 1 {

				// pass the value through, no trimming
				lines = append(lines, fmt.Sprintf("%s=%s", variable, data[1]))
			} else {
				// if only a pass-through variable is given, clean it up.
				lines = append(lines, fmt.Sprintf("%s=%s", strings.TrimSpace(line), os.Getenv(line)))
			}
		}
	}
	return lines, scanner.Err()
}

var whiteSpaces = " \t"

// ErrBadEnvVariable typed error for bad environment variable
type ErrBadEnvVariable struct {
	msg string
}

// Error implements error interface.
func (e ErrBadEnvVariable) Error() string {
	return fmt.Sprintf("poorly formatted environment: %s", e.msg)
}
