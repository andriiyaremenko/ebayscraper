package data

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/andriiyaremenko/ebayscraper/types"
)

type FileStorage interface {
	Flush() error
	Save(types.Product) error
}

func NewFileStorage(file string) FileStorage {
	return &fileStorage{fileName: file}
}

type fileStorage struct {
	mu       sync.Mutex
	fileName string
	items    []types.Product
}

func (fs *fileStorage) Save(p types.Product) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.items = append(fs.items, p)
	return nil
}

func (fs *fileStorage) Flush() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	f, err := os.Create(fs.fileName)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(fs.items)
	if err != nil {
		return err
	}
	_, err = f.Write(raw)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return err
	}
	fs.items = nil
	return nil
}
