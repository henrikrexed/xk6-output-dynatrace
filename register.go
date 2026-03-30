package dynatracewriter

import (
	"os"
	"strings"

	"github.com/Dynatrace/xk6-output-dynatrace/pkg/dynatracewriter"
	"go.k6.io/k6/output"
)

func init() {
	output.RegisterExtension("output-dynatrace", func(p output.Params) (output.Output, error) {
		mode := strings.ToLower(os.Getenv("K6_DYNATRACE_OUTPUT_MODE"))
		if mode == "otlp" {
			return dynatracewriter.NewOTLP(p)
		}
		return dynatracewriter.New(p)
	})
}
