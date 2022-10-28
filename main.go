package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mrjosh/udp2grpc/cmds"
	"github.com/spf13/cobra"
)

func main() {

	log.SetFlags(log.Lshortfile)

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
	rootCmd.AddCommand(cmds.NewServerCommand())
	rootCmd.AddCommand(cmds.NewClientCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
