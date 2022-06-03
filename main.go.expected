package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

const (
	VERSION          = "0.0.1"
	MAKEFILE         = "Makefile"
	MAKEFILE_MD      = "Makefile.md"
	README_MD        = "README.md"
	ENV_INPUT_FILE   = "MAKEDOWN_INPUT_FILE"
	ENV_OUTPUT_FILE  = "MAKEDOWN_OUTPUT_FILE"
	ENV_MAKE_COMMAND = "MAKEDOWN_MAKE_COMMAND"
	MAKE_COMMAND     = "make"
)

var (
	replace    bool
	verbose    bool
	inputfile  string
	outputfile string
	makepath   string
)

var rootCmd = &cobra.Command{
	Use:     "makedown [flags] [targets] ..",
	Version: VERSION,
	Short:   "'makedown' is a 'make' command wrapper for Markdown files.",
	Long: `'makedown' is a 'make' command wrapper for Markdown files.
You can write 'make' targets in 'Makefile.md' or 'README.md', etc.
'makedown' executes 'make' targets written in *.md files by the 'make' command.

For more information,
  https://github.com/hirokistring/makedown
`,
	Example: `  $ cat README.md

  # sayhello:
` + "    ```" + `
    @echo Hello, $(WHO)!
` + "    ```" + `
  # variables:
` + "    ```" + `
    WHO = makedown
` + "    ```" + `

  $ makedown sayhello
  Hello, makedown!
`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		MakeDownCommand(cmd, args)
	},
}

func MakeDownCommand(cmd *cobra.Command, args []string) {
	// setup logging configurations
	setupLogging()

	// Find the input markdown file.
	input_filename := findInputMarkdownFile()
	log.Printf("the input markdown file is %q\n", input_filename)

	// Read the input file.
	var md []byte
	var err error

	md, err = ioutil.ReadFile(input_filename)
	if err != nil {
		err = fmt.Errorf("failed to read bytes from %q: %v", input_filename, err)
		fmt.Fprintln(os.Stderr, err)
		log.Fatal(err)
	}

	// Generate the makefile contents from the input markdown file.
	out, _, err := GenerateMakefileFromMarkdown(input_filename, md)
	if err != nil {
		err = fmt.Errorf("failed to extract make targets from %q: %v", input_filename, err)
		fmt.Fprintln(os.Stderr, err)
		log.Fatal(err)
	}

	// Determine the output file name
	output_filename, keep := determineOutputFile()
	if exists(output_filename) {
		overwrite := askOverwrite(output_filename)
		if !overwrite {
			fmt.Println("makedown aborted.")
			os.Exit(0)
		}
	}

	// Write the makefile contents to the output file.
	log.Printf("generates the temporary makefile in %q\n", output_filename)
	err = ioutil.WriteFile(output_filename, out, 0644)
	if err != nil {
		err = fmt.Errorf("failed to write the temporary makefile in %q: %v", output_filename, err)
		fmt.Fprintln(os.Stderr, err)
		log.Fatal(err)
	}

	// Determine make command path
	makeCommandPath := determineMakeCommandPath()

	// Execute make targets
	if len(args) > 0 {
		log.Printf("executes command %q for %q with args: %s\n", makeCommandPath, output_filename, args)

		if isOutputFileGivenExplicitly() {
			// prepend the (-f) option to the make command
			args = append([]string{"-f", output_filename}, args...)
		}

		// Pass the rest of command line arguments to the 'make' command.
		cmd := exec.Command(makeCommandPath, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()

		if err != nil {
			err = fmt.Errorf("makedown: make command error: %v\nThe generated makefile is kept: %q\n", err, output_filename)
			fmt.Fprintln(os.Stderr, err)
			log.Fatal(err)
		}
	} else {
		// if no target is given, keep the generated makefile.
		keep = true
		log.Printf("no target is given. keep the generated makefile in %q", output_filename)
	}

	if keep {
		log.Printf("the temporary makefile is kept in %q", output_filename)
	} else {
		// Delete the temporary makefile.
		os.Remove(output_filename)
		log.Printf("deleted the temporary makefile: %q", output_filename)
	}
}

func setupLogging() {
	log.SetPrefix("[makedown] ")

	// Enable debug logs or not
	if verbose {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}
}

func findTargetsThatDoesNotExist(args []string, targets []string) {
	var norules []string
	for _, arg := range args {
		found := false
		for _, target := range targets {
			if arg+":" == target {
				found = true
			}
		}
		if !found {
			norules = append(norules, arg)
		}
	}
	if len(norules) > 0 {
		fmt.Fprintf(os.Stderr, "makedown: *** No rule to make target %q.\n", norules)
	}
}

func askOverwrite(output_filename string) bool {
	entered := false
	overwrite := false
	var answer string

	for i := 0; !entered && i < 3; i++ {
		fmt.Printf("File %q already exists. Overwrite? (y/n): ", output_filename)
		_, err := fmt.Scanln(&answer)
		if err == nil {
			switch answer {
			case "Y", "y", "yes", "Yes", "YES":
				overwrite = true
				entered = true
			case "N", "n", "no", "No", "NO":
				overwrite = false
				entered = true
			}
		}
	}

	log.Printf("File %q already exists. Overwrite? (y/n): %s\n", output_filename, strconv.FormatBool(overwrite))

	return overwrite
}

func determineCommandPath(command string) string {
	command_path, err := exec.LookPath(command)
	if err != nil {
		err = fmt.Errorf("command not found in the PATH: %q", command)
		fmt.Fprintln(os.Stderr, err)
		log.Fatal(err)
	}
	return command_path // resolved command path
}

func determineMakeCommandPath() string {
	if makepath != "" {
		return determineCommandPath(makepath)
	}

	// Otherwise, check if the environment variable is set.
	env_value, env_set := os.LookupEnv(ENV_MAKE_COMMAND)
	if env_set {
		return determineCommandPath(env_value)
	}

	// default make command
	return determineCommandPath(MAKE_COMMAND)
}

func isOutputFileGivenExplicitly() bool {
	_, env_set := os.LookupEnv(ENV_OUTPUT_FILE)
	return outputfile != "" || env_set
}

func determineOutputFile() (string, bool) {
	if outputfile != "" {
		// The output file is given explicitly by the command line option.
		return outputfile, true // keep
	}

	// Otherwise, check if the environment variable is set.
	env_value, env_set := os.LookupEnv(ENV_OUTPUT_FILE)
	if env_set {
		return env_value, true // keep
	}

	// Otherwise, use Makefile.md
	if exists(MAKEFILE) {
		// The output file will be kept, if it already exists.
		return MAKEFILE, true // keep
	}

	return MAKEFILE, false // temporary
}

func findInputMarkdownFile() string {
	if inputfile != "" {
		if exists(inputfile) {
			// The input file is specified explicitly as the command line argument.
			return inputfile
		} else {
			// Exit with error message.
			err := fmt.Errorf("the input markdown file is not found: %q", inputfile)
			fmt.Fprintln(os.Stderr, err)
			log.Fatal(err)
		}
	}

	// Otherwise, check if the environment variable is set.
	env_value, env_set := os.LookupEnv(ENV_INPUT_FILE)
	if env_set {
		if exists(env_value) {
			// Use the file given by MAKEDOWN_INPUT_FILE
			return env_value
		} else {
			// Just print a warning message.
			log.Printf("the input markdown file given by %q does not exist: %q\n", ENV_INPUT_FILE, env_value)
		}
	}

	// Otherwise, try to find Makefile.md at first.
	if exists(MAKEFILE_MD) {
		return MAKEFILE_MD
	}

	// Otherwise, try to find README.md next.
	if exists(README_MD) {
		return README_MD
	}

	// Otherwise, error.
	err := fmt.Errorf("no input markdown file is found.")
	fmt.Fprintln(os.Stderr, err)
	log.Fatal(err)
	return ""
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&inputfile, "markdown", "f", "", "input markdown file name")
	rootCmd.PersistentFlags().StringVarP(&outputfile, "out", "", "", "output makefile name")
	rootCmd.PersistentFlags().StringVarP(&makepath, "make", "", "", "make command name in the PATH, like 'gmake'")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "", false, "prints verbose messages")
}

func main() {
	// show help by default.
	if len(os.Args) < 2 {
		rootCmd.Help()
		os.Exit(0)
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		log.Fatal(err)
	}
}

// This file is generated from "main.go.md" by godown.
// https://github.com/hirokistring/godown
