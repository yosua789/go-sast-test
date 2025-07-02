package helper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
)

const (
	PublicDir = "public/"
	UploadDir = "upload/"
	LogoDir   = "logo/"
)

func CheckUploadDir(subpath string) error {
	checkPath := path.Join(PublicDir, UploadDir)
	if subpath != "" {
		checkPath = path.Join(checkPath, subpath)
	}
	_, err := os.Stat(checkPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			os.MkdirAll(checkPath, os.ModePerm)
		} else {
			return err
		}
	}
	return nil
}

func SaveUploadToPublic(filename string, buff bytes.Buffer) (filepath string, err error) {
	err = CheckUploadDir("")
	if err != nil {
		return
	}
	filepath = path.Join(PublicDir, UploadDir, filename)
	out, err := os.Create(filepath)
	if err != nil {
		return
	}
	defer out.Close()
	_, err = io.Copy(out, &buff)
	if err != nil {
		return
	}
	return
}

func SaveImage(subdir string, filename string, buff bytes.Buffer) (string, error) {
	err := CheckUploadDir(subdir)
	if err != nil {
		return "", err
	}
	filepath := path.Join(subdir, filename)
	filepath, err = SaveUploadToPublic(filepath, buff)
	if err != nil {
		return "", err
	}
	return filepath, err
}

func CopyFileToBuffer(file multipart.File) (*bytes.Buffer, error) {
	// Create a new bytes.Buffer
	buf := new(bytes.Buffer)

	// Copy the contents of the file into the buffer
	_, err := io.Copy(buf, file)
	if err != nil {
		return nil, fmt.Errorf("error copying file to buffer: %v", err)
	}

	return buf, nil
}
