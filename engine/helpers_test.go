package engine

import "os"

// errorString is a failure from outside the program, of the kind the system and
// asset helpers quote.
type errorString string

func (err errorString) Error() string {
	return string(err)
}

func writeFile(path string, contents string) error {
	return os.WriteFile(path, []byte(contents), 0o644)
}
