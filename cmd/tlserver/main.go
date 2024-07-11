package main

import (
	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/dmfr/importer"
	"github.com/interline-io/transitland-lib/dmfr/sync"
	"github.com/interline-io/transitland-lib/dmfr/unimporter"
	"github.com/interline-io/transitland-lib/extract"
	"github.com/interline-io/transitland-lib/merge"
	"github.com/interline-io/transitland-lib/tlcli"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: "tlserver"}

func init() {
	pc := "tlserver"
	rootCmd.AddCommand(
		tlcli.CobraHelper(&fetch.Command{}, pc, "fetch"),
		tlcli.CobraHelper(&sync.Command{}, pc, "sync"),
		tlcli.CobraHelper(&extract.Command{}, pc, "extract"),
		tlcli.CobraHelper(&fetch.RebuildStatsCommand{}, pc, "rebuild-stats"),
		tlcli.CobraHelper(&importer.Command{}, pc, "import"),
		tlcli.CobraHelper(&unimporter.Command{}, pc, "unimport"),
		tlcli.CobraHelper(&merge.Command{}, pc, "merge"),
		tlcli.CobraHelper(&Command{}, pc, "server"),
	)
}

func main() {
	rootCmd.Execute()
}
