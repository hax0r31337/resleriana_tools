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
    "strings"
)

type ConfigInfo struct {
	FetchUrl string `json:"fetch_url"`
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
	exeNameWithoutExt := strings.TrimSuffix(exeName, filepath.Ext(exeName))
	configFile := filepath.Join(exeNameWithoutExt + "_config.json") // Use filepath.Join for cross-platform path construction

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		var ConfigInfo ConfigInfo
		ConfigInfo.FetchUrl = "https://asset.resleriana.com/asset/AssetsVersion/Android/"

		file, err := os.Create(configFile)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		err = encoder.Encode(ConfigInfo)
		if err != nil {
			panic(err)
		}
	}

	var ConfigInfo ConfigInfo
	file, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&ConfigInfo)
	if err != nil {
		panic(err)
	}
	
	ConfigInfojsonData, err := json.MarshalIndent(&ConfigInfo, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("exeNameWithoutExt: %s\n", exeNameWithoutExt)
	fmt.Printf("configFile: %s\n", configFile)
	fmt.Printf("ConfigInfo: %+v\n", string(ConfigInfojsonData))

	BASE_URL = fmt.Sprintf("%s", ConfigInfo.FetchUrl)

	catalog := Catalog{}
	err = catalog.FetchCatalog()
	if err != nil {
		log.Printf("Failed to fetch catalog: %v\n", err)
	} else {
		fmt.Println("Catalog fetched successfully")
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
