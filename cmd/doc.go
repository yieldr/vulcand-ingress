package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var cmdDoc = &cobra.Command{
	Use:   "doc",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: runDoc,
}

func runDoc(cmd *cobra.Command, args []string) {

	err := doc.GenMarkdownTree(cmdRoot, "doc")
	if err != nil {
		fmt.Println(err)
		os.Exit(-2)
	}
}

func init() {
	cmdRoot.AddCommand(cmdDoc)
	cmdDoc.Flags().StringP("dir", "d", "doc", "directory in which to generate documentation")
}
