package exitchecker_test

import (
	"testing"

	"github.com/valinurovdenis/gophkeeper/internal/staticlint/exitchecker"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestMyAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), exitchecker.ExitCheckAnalyzer, "./...")
}
