package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mbrumlow/gopak/pak"
)

func main() {

	pak.Init()

	for _, f := range []string{
		"fileA.txt",
		"dir/fileB.txt"} {

		r, err := pak.Open("files", f)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("START[%v]\n", f)
		io.Copy(os.Stdout, r)
		fmt.Printf("END[%v]\n", f)

		r.Close()
	}

}
