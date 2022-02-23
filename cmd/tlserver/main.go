package main

import (
	"flag"
	"os"

	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/dmfr/importer"
	"github.com/interline-io/transitland-lib/dmfr/sync"
	"github.com/interline-io/transitland-lib/dmfr/unimporter"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-lib/tl"
	server "github.com/interline-io/transitland-server"
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
		log.Print("Usage of %s:", os.Args[0])
		log.Print("Commands:")
		log.Print("  sync")
		log.Print("  fetch")
		log.Print("  import")
		log.Print("  unimport")
		log.Print("  server")

	}
	flag.Parse()
	if versionFlag {
		log.Print("transitland-lib version: %s", tl.VERSION)
		log.Print("gtfs spec version: https://github.com/google/transit/blob/%s/gtfs/spec/en/reference.md", tl.GTFSVERSION)
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
	default:
		log.Print("%q is not valid command.", subc)
		return
	}
	if err := r.Parse(args[1:]); err != nil {
		log.Errorf("Error: %s", err.Error())
	}
	if err := r.Run(); err != nil {
		log.Errorf("Error: %s", err.Error())
	}
}
