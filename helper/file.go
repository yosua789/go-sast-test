package helper

import (
	"io"
	"os"
	"path"
	"strings"
)

func DeleteUploadFile(name string) bool {
	filepath := path.Join("./public/upload", name)
	_, err := os.Stat(filepath)
	if err != nil {
		return false
	}

	err = os.Remove(filepath)
	return err == nil
}

func GetFileExtension(fileName string) string {
	return fileName[strings.LastIndex(fileName, ".")+1:]
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func ReadFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 2. Baca semua isi file
	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
