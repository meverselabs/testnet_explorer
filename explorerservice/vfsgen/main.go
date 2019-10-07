package main

import (
	"log"
	"os"

	"github.com/fletaio/explorer/explorerservice"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(explorerservice.Assets, vfsgen.Options{
		PackageName:  "explorerservice",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatal(err)
	}

	oldLocation := "./assets_vfsdata.go"
	newLocation := "../assets_vfsdata.go"
	err = os.Rename(oldLocation, newLocation)
	if err != nil {
		log.Fatal(err)
	}
}
