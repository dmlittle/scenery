package parser

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreprocessPlan(t *testing.T) {
	cases := []struct {
		inputFile        string
		outputFile       string
		expectedWarnings []string
	}{
		{"../../fixtures/rawPlans/planOnlyInput.txt", "../../fixtures/rawPlans/planOnlyOutput.txt", nil},
		{"../../fixtures/rawPlans/ANSIInput.txt", "../../fixtures/rawPlans/ANSIOutput.txt", nil},
		{"../../fixtures/rawPlans/rulerInput.txt", "../../fixtures/rawPlans/rulerOutput.txt", nil},
		{"../../fixtures/rawPlans/preface1Input.txt", "../../fixtures/rawPlans/preface1Output.txt", nil},
		{"../../fixtures/rawPlans/preface2Input.txt", "../../fixtures/rawPlans/preface2Output.txt", nil},
		{"../../fixtures/rawPlans/preface3Input.txt", "../../fixtures/rawPlans/preface3Output.txt", nil},
		{"../../fixtures/rawPlans/noChangesInput.txt", "../../fixtures/rawPlans/noChangesOutput.txt", nil},
		{"../../fixtures/rawPlans/warningInput.txt", "../../fixtures/rawPlans/warningOutput.txt", []string{"Warning: test warning\n"}},
		{"../../fixtures/rawPlans/postfaceInput.txt", "../../fixtures/rawPlans/postfaceOutput.txt", nil},
	}

	for _, tc := range cases {
		input, err := ioutil.ReadFile(tc.inputFile)
		assert.NoError(t, err)

		expected, err := ioutil.ReadFile(tc.outputFile)
		assert.NoError(t, err)

		output, warnings := preprocessPlan(string(input))

		assert.Equal(t, string(expected), output)
		assert.Equal(t, tc.expectedWarnings, warnings)
	}
}

func TestParse(t *testing.T) {
	t.Run("parses simple plans", func(tt *testing.T) {
		input, err := ioutil.ReadFile("../../fixtures/processedPlans/simple.txt")
		assert.NoError(tt, err)

		expected := &Plan{
			Resources: []*Resource{
				{
					Header: &Header{
						Change: String("+"),
						Name:   String("module.module_name"),
					},
					Attributes: []*Attribute{
						{
							Key:      String("id"),
							Computed: String("<computed>"),
						},
					},
				},
			},
		}

		plan, err := Parse(string(input))
		assert.NoError(t, err)

		assert.Equal(tt, expected, plan)
	})

	t.Run("parses all types of resource changes", func(tt *testing.T) {
		input, err := ioutil.ReadFile("../../fixtures/processedPlans/resourceChangeTypes.txt")
		assert.NoError(tt, err)

		expected := &Plan{
			Resources: []*Resource{
				{
					Header: &Header{
						Change: String("+"),
						Name:   String("module.module_name"),
					},
				},
				{
					Header: &Header{
						Change: String("-"),
						Name:   String("module.module_name"),
					},
				},
				{
					Header: &Header{
						Change: String("-/+"),
						Name:   String("module.module_name"),
					},
				},
				{
					Header: &Header{
						Change: String("~"),
						Name:   String("module.module_name"),
					},
				},
				{
					Header: &Header{
						Change: String("<="),
						Name:   String("module.module_name"),
					},
				},
			},
		}

		plan, err := Parse(string(input))
		assert.NoError(t, err)

		assert.Equal(tt, expected, plan)
	})

	t.Run("parses plan metadata", func(tt *testing.T) {
		input, err := ioutil.ReadFile("../../fixtures/processedPlans/metadata.txt")
		assert.NoError(tt, err)

		expected := &Plan{
			Metadata: &Metadata{
				Add:     4,
				Change:  1,
				Destroy: 0,
			},
		}

		plan, err := Parse(string(input))
		assert.NoError(t, err)

		assert.Equal(tt, expected, plan)
	})

	t.Run("parses complex plan", func(tt *testing.T) {
		input, err := ioutil.ReadFile("../../fixtures/processedPlans/complex.txt")
		assert.NoError(tt, err)

		expected := &Plan{
			Resources: []*Resource{
				{
					Header: &Header{
						Change: String("+"),
						Name:   String("module.module_name"),
					},
					Attributes: []*Attribute{
						{
							Key:      String("id"),
							Computed: String("<computed>"),
						},
						{
							Key:   String("username"),
							Value: String("scenery"),
						},
						{
							Key:      String("password"),
							Computed: String("<sensitive>"),
						},
					},
				},
				{
					Header: &Header{
						Change: String("+"),
						Name:   String("module.module.name"),
					},
				},
				{
					Header: &Header{
						Change: String("+"),
						Name:   String("module.module-name"),
					},
				},
				{
					Header: &Header{
						Change: String("+"),
						Name:   String("module.module_name[0]"),
					},
				},
				{
					Header: &Header{
						Change: String("~"),
						Name:   String("module.module_name"),
					},
					Attributes: []*Attribute{
						{
							Key:    String("policy"),
							Before: String("{\n  \"Version\": \"1234\"\n}"),
							After:  String("{\n  \"Version\": \"5678\"\n}"),
						},
						{
							Key:           String("attribute"),
							Before:        String(""),
							AfterComputed: String("<computed>"),
						},
					},
				},
				{
					Header: &Header{
						Change:      String("-/+"),
						Name:        String("aws_instance.example"),
						NewResource: true,
					},
					Attributes: []*Attribute{
						{
							Key:         String("ami"),
							Before:      String("ami-2757f631"),
							After:       String("ami-b374d5a5"),
							NewResource: true,
						},
						{
							Key:   String("instance_tags.k8s.io/role/master"),
							Value: String("1"),
						},
						{
							Key:   String("k8s.io/cluster-autoscaler"),
							Value: String("nodes"),
						},
					},
				},
			}}

		plan, err := Parse(string(input))
		assert.NoError(t, err)

		assert.Equal(tt, expected, plan)
	})
}

func String(v string) *string {
	return &v
}
