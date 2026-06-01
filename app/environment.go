package app

import "strings"

// Environment provides read-only access to process environment values.
type Environment interface {
	LookupEnv(string) (string, bool)
}

// MapEnvironment is an immutable environment snapshot.
type MapEnvironment map[string]string

// NewEnvironment creates an environment snapshot from os.Environ-style values.
func NewEnvironment(values []string) MapEnvironment {
	env := make(MapEnvironment, len(values))
	for _, value := range values {
		name, val, ok := strings.Cut(value, "=")
		if !ok {
			continue
		}
		env[name] = val
	}

	return env
}

// LookupEnv returns the value for name.
func (e MapEnvironment) LookupEnv(name string) (string, bool) {
	value, ok := e[name]
	return value, ok
}
