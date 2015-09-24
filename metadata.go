package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var (
	recognizedOptions = []string{"share"}
)

type VolumeMetadata struct {
	CreatedAt time.Time     `json:"created_at"`
	Account   string        `json:"account"`
	Options   VolumeOptions `json:"options"`
}

type VolumeOptions struct {
	Share string `json:"share"`
}

type MetadataDriver struct {
	metaDir string
}

func NewMetadataDriver(metaDir string) (*MetadataDriver, error) {
	if err := os.MkdirAll(metaDir, 0700); err != nil {
		return nil, fmt.Errorf("error creating %s: %v", metaDir, err)
	}
	return &MetadataDriver{metaDir}, nil
}

func (m *MetadataDriver) Validate(meta map[string]string) (VolumeMetadata, error) {
	var v VolumeMetadata

	// Validate keys
	for k, _ := range meta {
		found := false
		for _, opts := range recognizedOptions {
			if k == opts {
				found = true
				break
			}
		}
		if !found {
			return v, fmt.Errorf("not a recognized volume driver option: %q", k)
		}
	}

	return VolumeMetadata{
		Options: VolumeOptions{
			Share: meta["share"]}}, nil
}

func (m *MetadataDriver) Set(name string, meta VolumeMetadata) error {
	b, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("cannot serialize metadata: %v", err)
	}
	if err := ioutil.WriteFile(m.path(name), b, 0600); err != nil {
		return fmt.Errorf("cannot write metadata: %v", err)
	}
	return nil
}

func (m *MetadataDriver) Get(name string) (VolumeMetadata, error) {
	var v VolumeMetadata
	b, err := ioutil.ReadFile(m.path(name))
	if err != nil {
		return v, fmt.Errorf("cannot read metadata: %v", err)
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return v, fmt.Errorf("cannot deserialize metadata: %v", err)
	}
	return v, nil
}

func (m *MetadataDriver) path(name string) string {
	return filepath.Join(m.metaDir, name)
}
