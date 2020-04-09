package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "ahisto",
	Short: "Reads a file or pipe of samtools depth -a output and calculates histogram of read depths. If line contains one number histogram will be calculated for that. ",
	Long: `From a simple demo of the usage of linux pipes
Transform the samtools depth input (pipe of file) to a histogram`,
	RunE: func(cmd *cobra.Command, args []string) error {
		print = logNoop
		if flags.verbose {
			print = logOut
		}
		return runCommand()
	},
}

var flags struct{
	filepath string
	verbose bool
}

var flagsName = struct{
	file, fileShort string
	verbose, verboseShort string
} {
	"file", "f",
	"verbose", "v",
}

var print func(s string)

func main() {
	rootCmd.Flags().StringVarP(
		&flags.filepath,
		flagsName.file,
		flagsName.fileShort,
		"", "path to the file")
	rootCmd.PersistentFlags().BoolVarP(
		&flags.verbose,
		flagsName.verbose,
		flagsName.verboseShort,
		false, "log verbose output")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func logNoop(s string) {}

func logOut(s string) {
	log.Println(s)
}
