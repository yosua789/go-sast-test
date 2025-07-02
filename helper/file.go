package helper

import (
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
