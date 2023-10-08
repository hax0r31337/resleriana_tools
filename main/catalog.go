package main

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
	Compression  uint16 `json:"_compression"`
}
