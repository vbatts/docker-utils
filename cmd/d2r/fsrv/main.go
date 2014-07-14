package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
)

var (
	root   = "./registry"
	flBind = flag.String("b", "127.0.0.1", "addr to bind to")
	flPort = flag.String("p", "5000", "port to listen on")
)

func main() {
	flag.Parse()

	if flag.NArg() > 0 {
		root = flag.Args()[0]
	}

	root, err := filepath.Abs(root)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.FileServer(http.Dir(root)))
	log.Printf("Serving %s on %s:%s ...", root, *flBind, *flPort)
	log.Fatal(http.ListenAndServe(*flBind+":"+*flPort, nil))
}
