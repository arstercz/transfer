package main

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
)

type Storage interface {
	Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error)
	Head(token string, filename string) (contentType string, contentLength uint64, err error)
	Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error
	Del(token string, filename string) (contentType string, err error)
}

type LocalStorage struct {
	Storage
	basedir string
}

func NewLocalStorage(basedir string) (*LocalStorage, error) {
	return &LocalStorage{basedir: basedir}, nil
}

func (s *LocalStorage) Head(token string, filename string) (contentType string, contentLength uint64, err error) {
	path := filepath.Join(s.basedir, token, filename)

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	contentLength = uint64(fi.Size())

	contentType = mime.TypeByExtension(filepath.Ext(filename))

	return
}

func (s *LocalStorage) Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error) {
	path := filepath.Join(s.basedir, token, filename)

	// content type , content length
	if reader, err = os.Open(path); err != nil {
		return
	}

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	contentLength = uint64(fi.Size())

	contentType = mime.TypeByExtension(filepath.Ext(filename))

	return
}

func (s *LocalStorage) Del(token string, filename string) (contentType string, err error) {
	path := filepath.Join(s.basedir, token, filename)

	_, err = os.Stat(path)
	if err != nil {
		return
	}
	contentType = mime.TypeByExtension(filepath.Ext(filename))

	err = os.Remove(path)
	if err != nil {
		return
	}
	return
}

func (s *LocalStorage) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error {
	var f io.WriteCloser
	var err error

	path := filepath.Join(s.basedir, token)

	if err = os.Mkdir(path, 0700); err != nil && !os.IsExist(err) {
		return err
	}

	if f, err = os.OpenFile(filepath.Join(path, filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		fmt.Printf("%s", err)
		return err
	}

	defer f.Close()

	if _, err = io.Copy(f, reader); err != nil {
		return err
	}

	return nil
}
