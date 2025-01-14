package cmd

import (
	"fmt"
	"log"
	"regexp"

	"github.com/spf13/cobra"
)

// proxyCmd proxies the stored routes
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start Traverser as a proxy between you and another API and save the responses locally",
	Long: `Start Traverser as a proxy between you and another API and save the responses locally. 
	Example: traverser proxy https://pokeapi.co/api/v2/pokemon.`,
	Args: cobra.MatchAll(cobra.ExactArgs(1)),
	Run: func(cmd *cobra.Command, args []string) {
		proxyURL := args[0]
		re := regexp.MustCompile(`^(http|https)://.+`)
		if !re.MatchString(proxyURL) {
			HandleError(fmt.Errorf("invalid URL: %s", proxyURL))
		}

		server.UsePort(port)
		server.RegisterProxyHandler(proxyURL)
		log.Printf("Started server on port: %d\n", port)
		err := server.Start()
		HandleError(err)
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)
	proxyCmd.Flags().IntVarP(&port, "port", "p", 4000, "Port")
}
