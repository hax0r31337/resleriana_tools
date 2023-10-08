package main

import (
	"aktsk/pack"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type task struct {
	r io.ReadCloser
	w io.WriteCloser
	k []byte
}

func main() {
	catalog := Catalog{}

	j, err := os.ReadFile("./dat/catalog.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(j, &catalog)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	taskc := make(chan task)
	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(taskc, &wg)
	}

	t := time.Now()

	for _, bundle := range catalog.Files.Bundles {
		b, err := os.Open("./dat/" + bundle.RelativePath)
		if err != nil {
			panic(err)
		}

		outPath := "./dat/out/" + bundle.RelativePath
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
			r: b,
			w: out,
			k: []byte(fmt.Sprintf("%s-%d-%s-%d", bundle.BundleName, bundle.FileSize-28, bundle.Hash, bundle.CRC)),
		}
		taskc <- task
		wg.Add(1)
	}

	wg.Wait()
	close(taskc)

	fmt.Println(time.Since(t))
}

func worker(c chan task, wg *sync.WaitGroup) {
	for t := range c {
		pack.ReadPackedAB(t.r, t.w, t.k)
		t.r.Close()
		t.w.Close()
		wg.Done()
	}
}
