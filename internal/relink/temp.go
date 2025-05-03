package relink

import "os"

func GetSafeTempFile(dir string, prefix string) (string, error) {
	tempFile, err := os.CreateTemp(dir, prefix)
	if err != nil {
		return "", err
	}
	err = os.Remove(tempFile.Name())
	if err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}
