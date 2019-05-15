package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dmlittle/scenery/pkg/parser"
	"github.com/dmlittle/scenery/pkg/printer"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	sceneryVersion string

	noColor bool
)

// Execute is the entrypoint of the CLI.
func Execute(version string) {
	sceneryVersion = version

	cmd := &cobra.Command{
		Use:     "scenery",
		Short:   "CLI for prettifying Terraform plan outputs",
		Example: "  terraform plan | scenery",
		Version: sceneryVersion,
		Run:     runScenery,
	}

	cmd.PersistentFlags().BoolVarP(&noColor, "no-color", "n", false, "Print output without color")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func printPlan(input string) {
	// If we have input
	if len(input) > 0 {
		plan, err := parser.Parse(input)
		if err != nil {
			if err == parser.ErrParseFailure {
				os.Stderr.WriteString(color.RedString("Failed to parse plan. Returning original input.\n")) // nolint: gosec
				fmt.Println(input)
				return
			}
		}

		// plan will be nil if the parser panicked (potentially due to unrecognized
		// character or sequences) so we return the original input.
		if plan == nil {
			os.Stderr.WriteString(color.RedString("Failed to parse plan. Returning original input.\n")) // nolint: gosec
			fmt.Println(input)
			return
		}

		printer.PrettyPrint(plan)
	}
}

func runScenery(cmd *cobra.Command, args []string) {
	var input string

	if noColor {
		color.NoColor = true
	}

	stat, _ := os.Stdin.Stat() // nolint: gosec

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		var lines []string
		scanner := bufio.NewScanner(os.Stdin)
		insidePlanBlock := false

		// Loop through input until separator
		for scanner.Scan() {
			text := scanner.Text()
			if strings.Contains(text, "------------------------------------------------------------------------") {
				if insidePlanBlock {
					// If we detect it again, its the end of the plan block
					insidePlanBlock = false
					input = strings.Join(lines, "\n")
					printPlan(input)
				} else {
					// Otherwise we are entering the plan block
					insidePlanBlock = true
				}
				fmt.Println(text)
			} else if insidePlanBlock {
				lines = append(lines, text)
			} else {
				fmt.Println(text)
			}
		}
	} else if len(args) == 1 {
		fileContents, err := ioutil.ReadFile(args[0])
		if err != nil {
			cmd.Usage() // nolint: gosec
			return
		}

		input = string(fileContents)
	} else {
		// If no stdin or arguments, print usage
		cmd.Usage() // nolint: gosec
		os.Exit(0)
	}

}
