package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "compile takes as input Sass or SCSS and outputs CSS",
	Long: `compile takes as input Sass or SCSS and outputs CSS

Usage: sass compile file.scss
`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("compile called")
	},
}

func init() {
	RootCmd.AddCommand(compileCmd)

	compileCmd.Flags().StringP("output", "o", "", "location of output CSS file")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compileCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
