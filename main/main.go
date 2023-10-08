package main

import (
	"aktsk/pack"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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

		err = pack.ReadPackedAB(b, out, []byte(fmt.Sprintf("%s-%d-%s-%d", bundle.BundleName, bundle.FileSize-28, bundle.Hash, bundle.CRC)))
		if err != nil {
			panic(err)
		}

		b.Close()
		out.Close()
	}
}
