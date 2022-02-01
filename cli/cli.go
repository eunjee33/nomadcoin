package cli

import (
	"flag"
	"fmt"
	"os"
	//"runtime"

	"github.com/eunjee33/nomadcoin/explorer"
	"github.com/eunjee33/nomadcoin/rest"
)

func usage() {
	fmt.Printf("Welcome to 노마드 코인\n\n")
	fmt.Printf("Please use the following flags:\n\n")
	fmt.Printf("-port=4000:		Set the PORT of the server\n")
	fmt.Printf("-mode=rest:		Choose between 'html' and 'rest'\n")
	os.Exit(0) // 모든 함수를 제거하지만, 그 전에 defer를 먼저 이행해줘
}

func Start() {

	if len(os.Args) == 1{
		usage()
	}

	port := flag.Int("port", 4000, "Set port of the server")
	mode := flag.String("mode", "rest", "Choose between 'html' and 'rest'")

	flag.Parse()

	switch *mode {
	case "rest":
		rest.Start(*port)
	case "html":
		explorer.Start(*port)
	default:
		usage()
	}
}