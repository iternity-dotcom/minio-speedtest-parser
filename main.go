package main

import (
	"fmt"
	"os"

	"github.com/iternity-dotcom/minio-speedtest-parser/speedtest"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println(fmt.Errorf("pass a single .json or .zip file, that contains a MinIO Speedtest result"))
		os.Exit(1)
	}

	inputFile := os.Args[1]
	r, err := speedtest.FromZipFile(inputFile)
	if err != nil {
		r, err = speedtest.FromJsonFile(inputFile)
	}

	if err != nil {
		fmt.Println(fmt.Errorf("cannot parse %s. This doesn't seem to be a MinIO Speetest result file", inputFile))
	}

	fmt.Println(r.String())

}
