package archiver

import "fmt"

func validateArchiveSelection(extensions []string) error {
	var invalidExtensions []string

	for _, extension := range extensions {
		if _, keyFound := extensionToContentType[extension]; !keyFound {
			invalidExtensions = append(invalidExtensions, extension)
		}
	}

	if len(invalidExtensions) == 0 {
		return nil
	}

	return fmt.Errorf("these extensions are not valid choices %v", invalidExtensions)
}

func validateExtension(extension string) error {
	if _, ok := extensionToContentType[extension]; !ok {
		return fmt.Errorf("A file format with extension %v is not supported", extension)
	}
	return nil
}
