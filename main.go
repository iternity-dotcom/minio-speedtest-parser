package main

import (
	"fmt"
	"os"

	"github.com/iternity-dotcom/minio-speedtest-parser/speedtest"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println(fmt.Errorf("pass a single .json file, that contains a MinIO Speedtest result"))
		os.Exit(1)
	}

	jsonFile := os.Args[1]
	r, err := speedtest.FromJsonFile(jsonFile)

	if err != nil {
		fmt.Println(fmt.Errorf("cannot parse %s. This doesn't seem to be a MinIO Speetest result file", jsonFile))
	}

	fmt.Println(r.String())

}
