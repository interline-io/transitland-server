package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/dmfr/importer"
	"github.com/interline-io/transitland-lib/dmfr/sync"
	"github.com/interline-io/transitland-lib/dmfr/unimporter"
	"github.com/interline-io/transitland-lib/tl"
	server "github.com/interline-io/transitland-server"
	"github.com/interline-io/transitland-server/workers"
)

///////////////

type runnable interface {
	Parse([]string) error
	Run() error
}

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
		log.Printf("  sync")
		log.Printf("  fetch")
		log.Printf("  import")
		log.Printf("  unimport")
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
	var r runnable
	switch subc {
	case "sync":
		r = &sync.Command{}
	case "import":
		r = &importer.Command{}
	case "unimport":
		r = &unimporter.Command{}
	case "fetch":
		r = &fetch.Command{}
	case "server":
		r = &server.Command{}
	case "workers":
		r = &workers.Command{Config: workers.Config{QueueName: "tlv2-default"}}
	default:
		fmt.Printf("%q is not valid command.", subc)
		return
	}
	if err := r.Parse(args[1:]); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	if err := r.Run(); err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
}
