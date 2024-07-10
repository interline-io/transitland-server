package main

import (
	"github.com/interline-io/transitland-lib/cmd/tlcli"
	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/dmfr/importer"
	"github.com/interline-io/transitland-lib/dmfr/sync"
	"github.com/interline-io/transitland-lib/dmfr/unimporter"
	"github.com/interline-io/transitland-lib/extract"
	"github.com/interline-io/transitland-lib/merge"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: "tlserver"}

func init() {
	rootCmd.AddCommand(
		tlcli.CobraHelper(&fetch.Command{}, "fetch"),
		tlcli.CobraHelper(&sync.Command{}, "sync"),
		tlcli.CobraHelper(&extract.Command{}, "extract"),
		tlcli.CobraHelper(&fetch.RebuildStatsCommand{}, "rebuild-stats"),
		tlcli.CobraHelper(&importer.Command{}, "import"),
		tlcli.CobraHelper(&unimporter.Command{}, "unimport"),
		tlcli.CobraHelper(&merge.Command{}, "merge"),
		tlcli.CobraHelper(&Command{}, "server"),
	)
}

func main() {
	rootCmd.Execute()
}
