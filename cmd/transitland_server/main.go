package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	dmfr "github.com/interline-io/transitland-lib/dmfr/cmd"
	"github.com/interline-io/transitland-lib/tl"
	server "github.com/interline-io/transitland-server"
)

///////////////

func main() {
	quietFlag := false
	debugFlag := false
	traceFlag := false
	versionFlag := false
	flag.BoolVar(&quietFlag, "q", false, "Only send critical errors to stderr")
	flag.BoolVar(&debugFlag, "v", false, "Enable verbose output")
	flag.BoolVar(&traceFlag, "vv", false, "Enable more verbose/query output")
	flag.BoolVar(&versionFlag, "version", false, "Show version and GTFS spec information")
	flag.Usage = func() {
		log.Printf("Usage of %s:", os.Args[0])
		log.Printf("Commands:")
		log.Printf("  dmfr")
		log.Printf("  server")
	}
	flag.Parse()
	if versionFlag {
		log.Printf("transitland-lib version: %s", tl.VERSION)
		log.Printf("gtfs spec version: https://github.com/google/transit/blob/%s/gtfs/spec/en/reference.md", tl.GTFSVERSION)
		return
	}
	args := flag.Args()
	subc := flag.Arg(0)
	if subc == "" {
		flag.Usage()
		return
	}
	type runnable interface {
		Run([]string) error
	}
	var r runnable
	var err error
	switch subc {
	case "dmfr":
		r = &dmfr.Command{}
	case "server":
		r = &server.Command{}
	default:
		fmt.Printf("%q is not valid command.", subc)
		return
	}
	err = r.Run(args[1:]) // consume first arg
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
}
