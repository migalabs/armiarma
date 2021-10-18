package main

import (
	"fmt"
	"log"
	"os"

	"github.com/protolambda/rumor/sh"
	"github.com/spf13/cobra"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	// For memmory profiling
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	mainCmd := cobra.Command{
		Use:   "rumor",
		Short: "Start Rumor",
	}
	mainCmd.AddCommand(sh.AttachCmd(), sh.BareCmd(), sh.FileCmd(), sh.ServeCmd(), sh.ShellCmd(), sh.ToolCmd())
	// Set extra logging level for Armiarma Additional logs
	//logrus.SetLevel(logrus.ErrorLevel)

	if err := mainCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to run Rumor: %v", err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
