package pluralsparser

import (
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

func (t *TestSuite) TestEvaluate_ValidExpressions() {
	for _, expression := range testExpressions {
		res, err := Evaluate(expression.Expression, expression.N)
		require.NoError(t.T(), err)
		require.Exactly(t.T(), expression.Truth, res)
	}
}

func (t *TestSuite) TestEvaluate_SyntaxError0() {
	_, err := Evaluate(")1>2", 0)
	require.Error(t.T(), err)
}

func (t *TestSuite) TestEvaluate_SyntaxError1() {
	_, err := Evaluate("01>2", 0)
	require.Error(t.T(), err)
}

func (t *TestSuite) TestEvaluate_SyntaxError2() {
	_, err := Evaluate("1>>2", 0)
	require.Error(t.T(), err)
}

func BenchmarkEvaluate(b *testing.B) {
	for idx, expression := range testExpressions {
		if idx > 4 {
			break
		}
		Evaluate(expression.Expression, expression.N)
	}
}
