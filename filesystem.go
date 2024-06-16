package main

import (
	"bytes"
	"os"
)

type File interface {
	Write(b []byte) (n int, err error)
}

type Filesystem interface {
	OpenFileRC(name string) (File, error)
}

type diskFs struct {
	Filesystem
}

// func (d *diskFs) OpenFileRC(name string) (billy.File, error) {
// 	return os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0666)
// }

func (d *diskFs) OpenFileAWC(name string) (File, error) {
	return os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
}

type memFile struct {
	*bytes.Buffer
}

func (m *memFile) Write(b []byte) (n int, err error) {
	return m.Buffer.Write(b)
}

type memFs struct {
	Filesystem

	files map[string][]byte
}

func newMemFs() *memFs {
	return &memFs{
		files: make(map[string][]byte),
	}
}

func (d *memFs) OpenFileRC(name string) (File, error) {
	return d.createOrGet(name), nil
}

func (d *memFs) OpenFileAWC(name string) (File, error) {
	return d.createOrGet(name), nil
}

func (d *memFs) createOrGet(name string) File {
	_, ok := d.files[name]
	if !ok {
		d.files[name] = []byte{}
	}

	return &memFile{
		Buffer: bytes.NewBuffer(d.files[name]),
	}
}
