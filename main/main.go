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
	r        io.ReadCloser
	w        io.WriteCloser
	k        []byte
	packMode uint8
}

func main() {
	catalog := Catalog{}
	err := catalog.FetchCatalog()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	taskc := make(chan task)
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

		task := task{
			r:        resp.Body,
			w:        out,
			k:        []byte(fmt.Sprintf("%s-%d-%s-%d", bundle.BundleName, bundle.FileSize-28, bundle.Hash, bundle.CRC)),
			packMode: bundle.Compression,
		}
		taskc <- task
		wg.Add(1)

		log.Printf("extracting %s\n", bundle.RelativePath)
	}

	wg.Wait()
	close(taskc)

	fmt.Println(time.Since(t))
}

func worker(c chan task, wg *sync.WaitGroup) {
	for t := range c {
		switch t.packMode {
		case PackModeAktsk:
			pack.ReadPackedAB(t.r, t.w, t.k)
		case PackModeNone:
			io.Copy(t.w, t.r)
		default:
			log.Fatalf("unknown pack mode: %d", t.packMode)
		}
		t.r.Close()
		t.w.Close()
		wg.Done()
	}
}
