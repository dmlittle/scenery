package printer

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/dmlittle/scenery/pkg/parser"

	"github.com/stretchr/testify/assert"
)

func TestPrintPlan(t *testing.T) {
	cases := []struct {
		inputFile  string
		outputFile string
	}{
		{"../../fixtures/rawPlans/base64Input.txt", "../../fixtures/rawPlans/base64Output.txt"},
		{"../../fixtures/rawPlans/base64CreateInput.txt", "../../fixtures/rawPlans/base64CreateOutput.txt"},
		{"../../fixtures/rawPlans/multilineAttributeInput.txt", "../../fixtures/rawPlans/multilineAttributeOutput.txt"},
		{"../../fixtures/rawPlans/floatInput.txt", "../../fixtures/rawPlans/floatOutput.txt"},
	}

	for _, tc := range cases {
		input, err := ioutil.ReadFile(tc.inputFile)
		assert.NoError(t, err)

		expected, err := ioutil.ReadFile(tc.outputFile)
		assert.NoError(t, err)

		plan, err := parser.Parse(string(input))
		assert.NoError(t, err)

		output := captureOutput(func() {
			PrettyPrint(plan)
		})

		assert.Equal(t, string(expected), output)
	}
}

// https://gist.github.com/hauxe/e935a7f9012bf2649710cf75af323dbf#file-output_capturing_full-go
func captureOutput(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
		log.SetOutput(os.Stderr)
	}()
	os.Stdout = writer
	os.Stderr = writer
	log.SetOutput(writer)
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, reader)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	writer.Close()
	return <-out
}
