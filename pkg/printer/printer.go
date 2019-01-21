package printer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/dmlittle/scenery/pkg/parser"
	"github.com/fatih/color"
	"github.com/pmezard/go-difflib/difflib"
)

var (
	attributeIndentation = strings.Repeat(" ", 4)

	greenSprintf  = color.New(color.FgGreen).SprintFunc()
	redSprintf    = color.New(color.FgRed).SprintFunc()
	yellowSprintf = color.New(color.FgYellow).SprintFunc()
)

// PrettyPrint prints the Plan to stdout
func PrettyPrint(p *parser.Plan) {
	if p.Warnings != nil {
		for _, w := range *p.Warnings {
			color.Yellow(w)
		}
		fmt.Println()
	}

	if p.NoChanges {
		fmt.Println("No changes.")
		return
	}

	for _, r := range p.Resources {
		printResource(r)
	}

	printMetadata(p.Metadata)
}

func printResource(r *parser.Resource) {
	c := getTypeColor(r.Header.Change)

	printHeader(r.Header, c)
	printAttributes(r.Attributes, c)
	fmt.Println()
}

func printHeader(header *parser.Header, printer *color.Color) {
	colorSprintf := printer.SprintFunc()

	fullName := *header.Name

	if header.Taint {
		fullName = fmt.Sprintf("%s (tainted)", fullName)
	}

	if header.NewResource {
		fullName = fmt.Sprintf("%s (new resource required)", fullName)
	}

	var changeSymbol string
	if *header.Change == "-/+" {
		changeSymbol = fmt.Sprintf("%s/%s", redSprintf("-"), greenSprintf("+"))
	} else {
		changeSymbol = colorSprintf(*header.Change)
	}

	fmt.Printf("%s %s\n", changeSymbol, colorSprintf(fullName))
}

func printAttributes(attributes []*parser.Attribute, printer *color.Color) {
	if len(attributes) == 0 {
		return
	}

	colorSprintf := printer.SprintFunc()

	var maxAttributeLength int

	for _, a := range attributes {
		if l := len(*a.Key); l > maxAttributeLength {
			maxAttributeLength = l
		}
	}

	// Account for the extra character taken my the colon (":") after the key name
	maxAttributeLength++

	for _, a := range attributes {
		if a.Computed != nil {
			printComputedAttribute(*a.Key, *a.Computed, maxAttributeLength, colorSprintf)
		} else if a.Value != nil {
			printSimpleAttribute(*a.Key, *a.Value, maxAttributeLength, colorSprintf)
		} else if a.AfterComputed != nil {
			printComplexAttribute(*a.Key, *a.Before, *a.AfterComputed, true, a.NewResource, maxAttributeLength)
		} else if a.Before != nil && a.After != nil {
			processComplexAttributes(a, maxAttributeLength)
		}
	}
}

func processComplexAttributes(a *parser.Attribute, indentLength int) {
	isBeforeReference := isTerraformReference(a.Before)
	isAfterReference := isTerraformReference(a.After)

	if *a.Before == *a.After {
		return
	}

	if isBeforeReference != isAfterReference || (isBeforeReference && isAfterReference) || *a.Before == "" || *a.After == "" {
		printComplexAttribute(*a.Key, *a.Before, *a.After, false, a.NewResource, indentLength)
	} else {
		bBefore := []byte(*a.Before)
		bAfter := []byte(*a.After)

		isBeforeJSON := json.Valid(bBefore) && ((*a.Before)[0] == '{' || (*a.Before)[0] == '[')
		isAfterJSON := json.Valid(bAfter) && ((*a.After)[0] == '{' || (*a.After)[0] == '[')

		var old, new interface{}

		json.Unmarshal(bBefore, &old) // nolint:gosec
		json.Unmarshal(bAfter, &new)  // nolint:gosec

		oldPretty, _ := json.MarshalIndent(old, "", "  ") // nolint:gosec
		newPretty, _ := json.MarshalIndent(new, "", "  ") // nolint:gosec

		if isBeforeJSON && isAfterJSON {
			diff := difflib.UnifiedDiff{
				A:       difflib.SplitLines(string(oldPretty)),
				B:       difflib.SplitLines(string(newPretty)),
				Context: 5,
			}
			diffText, _ := difflib.GetUnifiedDiffString(diff) // nolint: gosec

			printDiffAttribute(*a.Key, diffText, indentLength)
		} else {
			printComplexAttribute(*a.Key, *a.Before, *a.After, false, a.NewResource, indentLength)
		}
	}
}

func printComputedAttribute(key, value string, maxKeyLength int, printer func(a ...interface{}) string) {
	printModifier := fmt.Sprintf("%%s%%-%ds %%s\n", maxKeyLength)

	fmt.Printf(printModifier, attributeIndentation, fmt.Sprintf("%s:", key), printer(value))
}

func printSimpleAttribute(key, value string, maxKeyLength int, printer func(a ...interface{}) string) {
	printModifier := fmt.Sprintf("%%s%%-%ds \"%%s\"\n", maxKeyLength)

	formattedValue := formatValue(value, maxKeyLength)

	fmt.Printf(printModifier, attributeIndentation, fmt.Sprintf("%s:", key), printer(formattedValue))
}

func printComplexAttribute(key, before, after string, computed, newResource bool, maxKeyLength int) {
	var afterModifier, formattedAfterValue string
	if computed {
		afterModifier = "%s"
		formattedAfterValue = after
	} else {
		afterModifier = "\"%s\""
		formattedAfterValue = formatValue(after, maxKeyLength)
	}

	formattedBeforeValue := formatValue(before, maxKeyLength)

	printModifier := fmt.Sprintf("%%s%%-%ds \"%%s\" => %s %%s\n", maxKeyLength, afterModifier)

	resourceText := ""
	if newResource {
		resourceText = yellowSprintf("(forces new resource)")
	}

	fmt.Printf(printModifier, attributeIndentation, fmt.Sprintf("%s:", key), redSprintf(formattedBeforeValue), greenSprintf(formattedAfterValue), resourceText)
}

func printDiffAttribute(key, diff string, maxKeyLength int) {
	printModifier := fmt.Sprintf("%%s%%-%ds ", maxKeyLength)

	fmt.Printf(printModifier, attributeIndentation, fmt.Sprintf("%s:", key))

	// 4 (attribute padding) + 1 (key/value space separation)
	diffIdentLength := maxKeyLength + 4 + 1
	diffPadding := strings.Repeat(" ", diffIdentLength)

	var padding string
	printedFirstLine := true
	lines := strings.Split(diff, "\n")
	for i, l := range lines {
		// Skip diff control lines (@@ -132,8 +134,8 @@)
		if strings.HasPrefix(l, "@@") && strings.HasSuffix(l, "@@") {
			continue
		}

		if printedFirstLine {
			padding = ""
			printedFirstLine = false
		} else {
			padding = diffPadding
		}

		if len(l) > 0 && l[0] == '+' {
			fmt.Printf("%s%s", padding, greenSprintf(l))
		} else if len(l) > 0 && l[0] == '-' {
			fmt.Printf("%s%s", padding, redSprintf(l))
		} else {
			fmt.Print(padding, l)
		}

		if i < len(lines) {
			fmt.Println()
		}
	}
}

func formatValue(value string, indentLength int) string {
	if json.Valid([]byte(value)) {
		// Is JSON?
		var j interface{}

		json.Unmarshal([]byte(value), &j) // nolint:gosec

		// 4 (attribute padding) + 1 (key/value space separation) + 1 (opening quote for value ")
		jsonIdentLegth := indentLength + 4 + 1 + 1

		formattedValue, _ := json.MarshalIndent(j, strings.Repeat(" ", jsonIdentLegth), "  ") // nolint:gosec

		return string(formattedValue)
	} else if strings.Contains(value, "\n") {
		// Is multi-line value?

		// 4 (attribute padding) + 1 (key/value space separation) + 1 (opening quote for value ")
		multiIdentLength := indentLength + 4 + 1 + 1

		newlineReplacement := fmt.Sprintf("\n%s", strings.Repeat(" ", multiIdentLength))

		formattedValue := strings.Replace(value, "\n", newlineReplacement, -1)

		lastNewlineIndex := strings.LastIndex(formattedValue, "\n")

		return formattedValue[:lastNewlineIndex]
	}

	return value
}

func printMetadata(metadata *parser.Metadata) {
	if metadata != nil {
		var add, change, destroy string

		if metadata.Add > 0 {
			add = color.New(color.FgGreen).SprintFunc()(fmt.Sprintf("%d to add", metadata.Add))
		} else {
			add = fmt.Sprintf("0 to add")
		}

		if metadata.Change > 0 {
			change = color.New(color.FgYellow).SprintFunc()(fmt.Sprintf("%d to change", metadata.Change))
		} else {
			change = fmt.Sprintf("0 to change")
		}

		if metadata.Destroy > 0 {
			destroy = color.New(color.FgRed).SprintFunc()(fmt.Sprintf("%d to destroy", metadata.Destroy))
		} else {
			destroy = fmt.Sprintf("0 to destroy")
		}

		fmt.Printf("Plan: %s, %s, %s.\n", add, change, destroy)
	}
}

func isTerraformReference(s *string) bool {
	terraformReference := regexp.MustCompile(`\$\{[a-zA-Z-_\.]+\}`)
	return terraformReference.MatchString(*s)
}

func getTypeColor(c *string) *color.Color {
	switch *c {
	case "+":
		return color.New(color.FgGreen)
	case "-":
		return color.New(color.FgRed)
	case "~", "-/+":
		return color.New(color.FgYellow)
	case "<=":
		return color.New(color.FgCyan)
	}
	return color.New(color.Reset)
}
