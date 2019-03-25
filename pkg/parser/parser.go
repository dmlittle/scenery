package parser

import (
	"errors"
	"regexp"

	"github.com/alecthomas/participle"
)

// The Plan struct is the root of the AST grammar used to parse the
// Terraform plan output.
type Plan struct {
	Warnings  *[]string
	_         *string     `parser:"{\"\\n\"}"`
	Resources []*Resource `parser:"{@@}"`
	_         *string     `parser:"{\"\\n\"}"`
	Metadata  *Metadata   `parser:"{@@}"`
	_         *string     `parser:"{\"\\n\"}"`

	NoChanges bool
}

// The Metadata struct is responsible for parsing the plan metadata that
// displays the summary statistics from the Terraform plan output.
//
// Example:
//   `Plan: 2 to add, 0 to change, 0 to destroy.`
type Metadata struct {
	_       *string `parser:"\"Plan\" \":\""`
	Add     int     `parser:"@Int \"to\" \"add\" \",\""`
	Change  int     `parser:"@Int \"to\" \"change\" \",\" "`
	Destroy int     `parser:"@Int \"to\" \"destroy\" \".\" "`
	_       *string `parser:"{\"\\n\"}"`
}

// The Resource struct is responsible for parsing each resource group that is
// displayed by the Terraform plan output. A resource comprises of a single
// Header and as many optional Attributes.
type Resource struct {
	Header     *Header      `parser:"@@"`
	Attributes []*Attribute `parser:"{ @@ }"`
	_          *string      `parser:"{ \"\\n\" }"`
}

// The Header struct is responsible for parsing the header of each resource
// displayed by Terraform plan output.
//
// Examples:
//   `+ aws_route53_record.record`
//   `-/+ module.module_name (new resource required)`
type Header struct {
	Change      *string `parser:"@(\"-\" \"/\" \"+\" | \"<\" \"=\" | \"+\" | \"-\" | \"~\")"`
	Name        *string `parser:"@(Ident { (\".\" | \"-\") (Ident | Int)+ | \"[\" Int \"]\" })"`
	Taint       bool    `parser:"{ @(\"(\" \"tainted\" \")\") }"`
	NewResource bool    `parser:"{ @(\"(\" \"new\" \"resource\" \"required\" \")\") }"`
	_           *string `parser:"\"\\n\""`
}

// The Attribute struct is responsible for parsing the attributes of each
// resource displayed by Terraform plan output.
//
// Examples:
//   `id:              <computed>`
//   `policy_arn:      "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"``
//   `allow_overwrite: "" => "true"`
type Attribute struct {
	Key           *string `parser:"@(Ident { \".\" | \"#\" | \"%\" | \"~\" | \"/\" | \"-\" | Ident | Float }) \":\""`
	Before        *string `parser:"((@String \"=\" \">\""`
	After         *string `parser:"  (@String"`
	AfterComputed *string `parser:" 	| @(\"<\" Ident \">\")))"`
	Value         *string `parser:" | @String"`
	Computed      *string `parser:" | @(\"<\" Ident \">\"))"`
	NewResource   bool    `parser:"{ @(\"(\" \"forces\" \"new\" \"resource\" \")\") }"`
	_             *string `parser:"\"\\n\""`
}

const noChanges = "NO_CHANGES_STRING"

// ErrParseFailure is returned by parser.Parse when the input string is unable
// to be parsed.
var ErrParseFailure = errors.New("validator not registered")

// Parse takes in an Terraform plan output string and returns a parsed
// representation in the form of a Plan struct.
//
// The ParseFailureErr error is returned the string is not able to be parsed
// properly by the grammar.
func Parse(inputPlan string) (*Plan, error) {
	defer func() {
		// Parse will panic in the event of unrecognized character sequences or
		// unsupported tokens. If we cannot parse the input it means it's not a
		// valid terraform plan. We'll recover and return the original input.
		recover()
	}()

	p, err := participle.Build(&Plan{}, participle.Lexer(&SceneryDefinition{}))
	if err != nil {
		return &Plan{}, err
	}

	processedPlan, warnings := preprocessPlan(inputPlan)

	if processedPlan == noChanges {
		return &Plan{
			NoChanges: true,
			Warnings:  &warnings,
		}, nil
	}

	plan := &Plan{}

	err = p.ParseString(processedPlan, plan)
	if err != nil {
		return nil, ErrParseFailure
	}

	if warnings != nil {
		plan.Warnings = &warnings
	}

	return plan, nil
}

func preprocessPlan(planText string) (string, []string) {
	var warnings []string
	processedPlanText := planText

	// Strip ANSI escape codes
	ansiRE := regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")
	processedPlanText = ansiRE.ReplaceAllString(processedPlanText, "")

	// Strip Terraform initialization messages. These preface messages
	// are not currently handled by the parser.
	//
	// Example:
	// 		"- Downloading plugin for provider "aws" (1.1.0)..."
	initRE := regexp.MustCompile("- .*\\.\\.\\.")
	processedPlanText = initRE.ReplaceAllString(processedPlanText, "")

	// Strip Terraform section separators ("--------...")
	separatorRE := regexp.MustCompile("--------+")
	processedPlanText = separatorRE.ReplaceAllString(processedPlanText, "")

	// Process Warnings
	warningRE := regexp.MustCompile("Warning:.*\n")
	if w := warningRE.FindStringSubmatch(processedPlanText); len(w) > 0 {
		warnings = w
		processedPlanText = warningRE.ReplaceAllString(processedPlanText, "")
	}

	// Strip preface
	pathRE := regexp.MustCompile("Path:[^\n]+\n")
	actionsRE := regexp.MustCompile("Terraform will perform the following actions:.*\n")
	noopPlanRE := regexp.MustCompile("(No changes|This plan does nothing).*\n")
	switch {
	case pathRE.MatchString(processedPlanText):
		matches := pathRE.FindAllStringIndex(processedPlanText, -1)
		lastMatchEndIndex := matches[len(matches)-1][1]
		processedPlanText = processedPlanText[lastMatchEndIndex:]
	case actionsRE.MatchString(processedPlanText):
		matches := actionsRE.FindAllStringIndex(processedPlanText, -1)
		lastMatchEndIndex := matches[len(matches)-1][1]
		processedPlanText = processedPlanText[lastMatchEndIndex:]
	case noopPlanRE.MatchString(processedPlanText):
		return noChanges, warnings
	}

	// Strip postface
	planRE := regexp.MustCompile("Plan:[^\n]+")
	if planRE.MatchString(processedPlanText) {
		matches := planRE.FindAllStringIndex(processedPlanText, -1)
		lastMatchEndIndex := matches[len(matches)-1][1]
		processedPlanText = processedPlanText[:lastMatchEndIndex]
	}

	return processedPlanText, warnings
}
