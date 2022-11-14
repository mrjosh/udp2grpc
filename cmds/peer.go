package cmds

import (
	"fmt"
	"strings"

	"github.com/mrjosh/udp2grpc/internal/config"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

type NewPeerCommandFlags struct {
	name           string
	remoteaddr     string
	available_from []string
}

func newPeerCommand() *cobra.Command {
	cFlags := new(NewPeerCommandFlags)
	cmd := &cobra.Command{
		Use:   "peer",
		Short: "Create a new peer",
		RunE: func(cmd *cobra.Command, args []string) error {

			if cFlags.name == "" {
				return fmt.Errorf("name is required")
			}

			if cFlags.remoteaddr == "" {
				return fmt.Errorf("remote is required")
			}

			if len(cFlags.available_from) == 0 {
				cFlags.available_from = append(cFlags.available_from, "0.0.0.0/0")
			}

			pk := generatePrivateKey()
			peer := []*config.PeerConfMap{
				{
					Name:          cFlags.name,
					PrivateKey:    pk.String(),
					Remote:        cFlags.remoteaddr,
					AvailableFrom: cFlags.available_from,
				},
			}

			serverOut, err := yaml.Marshal(peer)
			if err != nil {
				return err
			}

			fmt.Println("server side config:")
			fmt.Println("-------------------------------------------------------------")
			fmt.Println("...")
			fmt.Println("peers: \n" + strings.TrimSpace(string(serverOut)) + "\n...\n")

			clientConf := struct {
				Client *config.ClientConfMap `yaml:"client"`
			}{
				Client: &config.ClientConfMap{
					PrivateKey: pk.String(),
					Remote:     "{{ server ip address }}",
				},
			}

			clientOut, err := yaml.Marshal(clientConf)
			if err != nil {
				return err
			}

			fmt.Println("client side config:")
			fmt.Println("-------------------------------------------------------------")
			fmt.Println(strings.TrimSpace(string(clientOut)) + "\n  ...\n")
			return err
		},
	}
	cmd.SuggestionsMinimumDistance = 1
	cmd.Flags().StringVarP(&cFlags.name, "name", "n", "", "Peer name")
	cmd.Flags().StringVarP(&cFlags.remoteaddr, "remote", "r", "", "Remote addr")
	cmd.Flags().StringArrayVarP(&cFlags.available_from, "available-from", "A", []string{}, "available from ip/subnet")
	return cmd
}
