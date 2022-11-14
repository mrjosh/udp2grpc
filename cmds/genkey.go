package cmds

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mrjosh/udp2grpc/internal/protocol"
	"github.com/spf13/cobra"
)

func newGenKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genkey",
		Short: "generate a new privatekey",
		RunE: func(cmd *cobra.Command, args []string) error {
			pk := generatePrivateKey()
			fmt.Fprintln(cmd.OutOrStderr(), pk.String())
			return nil
		},
	}
	cmd.SuggestionsMinimumDistance = 1
	return cmd
}

func generatePrivateKey() *protocol.NoisePrivateKey {
	noisePK := &protocol.NoisePrivateKey{}
	rand.Seed(time.Now().UnixNano())
	min, max := 10, 99
	for i := 0; i < protocol.NoisePrivateKeySize; i++ {
		noisePK[i] = byte(rand.Intn(max-min) + min)
	}
	return noisePK
}
