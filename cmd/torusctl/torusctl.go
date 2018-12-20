package main

import (
	"fmt"
	"os"

	"github.com/alternative-storage/torus"
	"github.com/alternative-storage/torus/internal/flagconfig"
	"github.com/coreos/pkg/capnslog"
	"github.com/spf13/cobra"
)

var (
	logpkg string
	debug  bool
)

var rootCommand = &cobra.Command{
	Use:              "torusctl",
	Short:            "Administer the torus storage cluster",
	Long:             `Admin utility for the torus distributed storage cluster.`,
	PersistentPreRun: configure,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
		os.Exit(1)
	},
}

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("torusctl\nVersion: %s\n", torus.Version)
		os.Exit(0)
	},
}

func init() {
	rootCommand.PersistentFlags().BoolVarP(&debug, "debug", "", false, "enable debug logging")
	rootCommand.PersistentFlags().StringVarP(&logpkg, "logpkg", "", "", "Specific package logging")
	rootCommand.AddCommand(initCommand)
	rootCommand.AddCommand(blockCommand)
	rootCommand.AddCommand(listPeersCommand)
	rootCommand.AddCommand(ringCommand)
	rootCommand.AddCommand(peerCommand)
	rootCommand.AddCommand(volumeCommand)
	rootCommand.AddCommand(versionCommand)
	rootCommand.AddCommand(wipeCommand)
	rootCommand.AddCommand(configCommand)
	rootCommand.AddCommand(completionCommand)
	flagconfig.AddConfigFlags(rootCommand.PersistentFlags())
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		die("%v", err)
	}
}

func configure(cmd *cobra.Command, args []string) {
	capnslog.SetGlobalLogLevel(capnslog.WARNING)

	if debug {
		capnslog.SetGlobalLogLevel(capnslog.DEBUG)
	}
	if logpkg != "" {
		capnslog.SetGlobalLogLevel(capnslog.NOTICE)
		rl := capnslog.MustRepoLogger("github.com/alternative-storage/torus")
		llc, err := rl.ParseLogLevelConfig(logpkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing logpkg: %s\n", err)
			os.Exit(1)
		}
		rl.SetLogLevel(llc)
	}
}
