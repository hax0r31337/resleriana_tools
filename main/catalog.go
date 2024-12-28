package main

import (
	"encoding/json"
	"io"
	"fmt"
	"log"
)

const (
	PackModeNone  = 0
	PackModeAktsk = 3
)

type Catalog struct {
	Files struct {
		Bundles []BundleData `json:"_bundles"`
	} `json:"_fileCatalog"`
}

type BundleData struct {
	RelativePath string `json:"_relativePath"`
	BundleName   string `json:"_bundleName"`
	Hash         string `json:"_hash"`
	CRC          uint64 `json:"_crc"`
	FileSize     uint64 `json:"_fileSize"`
	FileMd5      string `json:"_fileMd5"`
	Compression  uint8  `json:"_compression"`
}

func (c *Catalog) FetchCatalog() error {
	resp, err := fetch("/catalog.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	j, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	err = json.Unmarshal(j, c)
	if err != nil {
		fmt.Printf("FetchCatalog data: %s\n", string(j))
		log.Println("Unmarshal error:", err) 
		return err 
	} else {
		return json.Unmarshal(j, c)
	}
}
