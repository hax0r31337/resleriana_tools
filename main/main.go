package main

import (
	"aktsk/pack"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const BASE_URL = "https://asset.resleriana.jp/asset/{REPLACE THIS WITH ASSET VERSION}/Android/"

type task struct {
	reader   io.ReadCloser
	writer   io.WriteCloser
	key      []byte
	packMode uint8
}

func main() {
	catalog := Catalog{}
	err := catalog.FetchCatalog()
	if err != nil {
		panic(err)
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

		outPath := "./extracted/" + bundle.RelativePath
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
			reader:   resp.Body,
			writer:   out,
			key:      []byte(fmt.Sprintf("%s-%d-%s-%d", bundle.BundleName, bundle.FileSize-28, bundle.Hash, bundle.CRC)),
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
