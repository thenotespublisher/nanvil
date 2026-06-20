package compiler

import (
	"github.com/nspcc-dev/neo-go/pkg/nsmith/detect"
)

// DetectLanguage resolves language and primary path from user input.
func DetectLanguage(path string) (Language, string, error) {
	res, err := detect.Detect(path)
	if err != nil {
		return "", "", err
	}
	return Language(res.Language), res.Path, nil
}
