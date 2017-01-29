package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mbrumlow/gopak/pak"
)

var root = flag.String("root", ".", "Root directory.")
var file = flag.String("file", "", "Go program to append pak.")

type pakFooter struct {
	Namespace string
	Size      int64
}

func main() {

	flag.Parse()

	if *file == "" {
		fmt.Printf("Please specify -file to append pak.\n")
		flag.Usage()
	}

	r := filepath.Clean(*root)

	out, _ := os.OpenFile(*file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	pk, _ := pak.NewPackWriter(r, out)

	filepath.Walk(r,
		func(path string, info os.FileInfo, err error) error {

			fi, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("Failed to stat file: %v", err)
			}

			if fi.Mode().IsDir() {
				return nil
			}

			if err := pk.AddFile(path); err != nil {
				log.Fatal(err)
			}

			return nil
		})

	if err := pk.Close(); err != nil {
		log.Fatal(err)
	}

	out.Close()
}
