package main

import (
	"aktsk/pack"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type VersionInfo struct {
	AssetsVersion string `json:"assets_version"`
}

var BASE_URL string

type task struct {
	reader  io.ReadCloser
	writer  io.WriteCloser
	key     []byte
	packMode uint8
}

func main() {
	exeName := filepath.Base(os.Args[0])
	versionFile := filepath.Join(exeName + "_versions.json") // Use filepath.Join for cross-platform path construction

	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		var versionInfo VersionInfo
		versionInfo.AssetsVersion = "AssetsVersion"

		file, err := os.Create(versionFile)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		err = encoder.Encode(versionInfo)
		if err != nil {
			panic(err)
		}
	}

	var versionInfo VersionInfo
	file, err := os.Open(versionFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&versionInfo)
	if err != nil {
		panic(err)
	}

	fmt.Printf("exeName: %s\n", exeName)
	fmt.Printf("versionFile: %s\n", versionFile)
	fmt.Printf("versionInfo: %+v\n", versionInfo)

	BASE_URL = fmt.Sprintf("https://asset.resleriana.com/asset/%s/Android/", versionInfo.AssetsVersion)

	catalog := Catalog{}
	err = catalog.FetchCatalog()
	if err != nil {
		fmt.Printf("version \"%+v\" is not Incorrect.\n", versionInfo.AssetsVersion)
	}

	var wg sync.WaitGroup
	taskc := make(chan *task)
	for i := 0; i < 16; i++ {
		go worker(taskc, &wg)
	}

	t := time.Now()

	for _, bundle := range catalog.Files.Bundles {
		resp, err := fetch(bundle.RelativePath)
		if err != nil {
			panic(err)
		}

		outPath := filepath.Join("./extracted/", bundle.RelativePath) // Use filepath.Join for path construction
		outDir := filepath.Dir(outPath)
		if _, err := os.Stat(outDir); os.IsNotExist(err) {
			err = os.MkdirAll(outDir, 0755)
			if err != nil {
				panic(err)
			}
		}

		out, err := os.Create(outPath)
		if err != nil {
			panic(err)
		}

		task := &task{
			reader:  resp.Body,
			writer:  out,
			key:     []byte(fmt.Sprintf("%s-%d-%s-%d", bundle.BundleName, bundle.FileSize-28, bundle.Hash, bundle.CRC)),
			packMode: bundle.Compression,
		}

		log.Printf("extracting %s\n", bundle.RelativePath)
		wg.Add(1)
		taskc <- task
	}

	wg.Wait()
	close(taskc)

	fmt.Println(time.Since(t))
}

func worker(c <-chan *task, wg *sync.WaitGroup) {
	var err error
	for t := range c {
		switch t.packMode {
		case PackModeAktsk:
			err = pack.ReadPackedAB(t.reader, t.writer, t.key)
		case PackModeNone:
			_, err = io.Copy(t.writer, t.reader)
		default:
			log.Fatalf("unknown pack mode: %d", t.packMode)
		}
		if err != nil {
			panic(err)
		}
		t.reader.Close()
		t.writer.Close()
		wg.Done()
	}
}
