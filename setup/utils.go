package setup

import (
	"bytes"
	"os"
	"path/filepath"
)

func ReplaceInFile(path, old, new string) (err error) {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if !bytes.Contains(content, []byte(old)) {
		return nil
	}

	newContent := bytes.ReplaceAll(content, []byte(old), []byte(new))

	tempFile, err := os.CreateTemp(filepath.Dir(path), "replace-")
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
		}
	}()

	_, err = tempFile.Write(newContent)
	if err != nil {
		return err
	}

	err = tempFile.Close()
	if err != nil {
		return err
	}

	err = os.Chmod(tempFile.Name(), info.Mode())
	if err != nil {
		return err
	}

	err = os.Rename(tempFile.Name(), path)
	return err
}
