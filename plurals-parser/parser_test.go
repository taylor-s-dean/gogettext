//go:generate goyacc -o parser.go parser.yy
package pluralsparser

import (
	// "fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
}

func TestPluralsParser(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (t *TestSuite) TestEvaluate() {
	for _, expression := range testExpressions {
		require.Exactly(t.T(), expression.Truth, Evaluate(expression.Expression, expression.N))
	}
}

func BenchmarkEvaluate(b *testing.B) {
	for _, expression := range testExpressions {
		Evaluate(expression.Expression, expression.N)
	}
}
