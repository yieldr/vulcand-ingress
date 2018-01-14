package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var cmdDoc = &cobra.Command{
	Use:   "doc",
	Short: "Generates cli documentation",
	Run:   runDoc,
}

func runDoc(cmd *cobra.Command, args []string) {

	dir, _ := cmd.Flags().GetString("dir")
	format, _ := cmd.Flags().GetString("format")

	var err error

	switch format {
	case "markdown":
		err = doc.GenMarkdownTree(cmdRoot, dir)
	case "rest":
		err = doc.GenReSTTree(cmdRoot, dir)
	case "man":
		err = doc.GenManTree(cmdRoot, &doc.GenManHeader{}, dir)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cmdRoot.AddCommand(cmdDoc)
	cmdDoc.Flags().StringP("dir", "d", "doc", "Directory in which to generate documentation")
	cmdDoc.Flags().StringP("format", "f", "markdown", "Format in which to generate documentation")
}
