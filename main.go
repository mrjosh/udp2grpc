package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mrjosh/udp2grpc/cmds"
	"github.com/mrjosh/udp2grpc/internal/version"
	"github.com/spf13/cobra"
)

var (
	BranchName string
	Version    string
	CompiledBy string
	BuildTime  string
)

func main() {
	rootCmd := &cobra.Command{
		Use: "utg",
		Long: `
   __  ______  ____ ___         ____  ____  ______
  / / / / __ \/ __ \__ \ ____  / __ \/ __ \/ ____/
 / / / / / / / /_/ /_/ // __ \/ /_/ / /_/ / /     
/ /_/ / /_/ / ____/ __// /_/ / _, _/ ____/ /___   
\____/_____/_/   /____/\__, /_/ |_/_/    \____/   
                      /____/`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.SetArgs(os.Args[1:])

	vi := &version.BuildInfo{
		Version:    Version,
		Branch:     BranchName,
		CompiledBy: CompiledBy,
		GoVersion:  runtime.Version(),
		BuildTime:  BuildTime,
	}

	if err := cmds.Start(vi, rootCmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
