package pluralsparser

import (
	"fmt"
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

	for _, expression := range []testExpression{
		{Expression: "1 + n", N: 1, Truth: 2},
		{Expression: "1 - n", N: 1, Truth: 0},
		{Expression: "1 * n", N: 1, Truth: 1},
		{Expression: "2 / n", N: 2, Truth: 1},
	} {
		res, err := Evaluate(expression.Expression, expression.N)
		require.NoError(t.T(), err)
		require.Exactly(t.T(), expression.Truth, res)
	}

}

func ExampleEvaluate() {
	result, err := Evaluate("n % 10 == 1 ? 1 : 0", 11)
	if err != nil {
		fmt.Printf("result = %d, err = %s", result, err)
	} else {
		fmt.Printf("result = %d, err = nil", result)
	}
	// Output: result = 1, err = nil
}

func (t *TestSuite) TestEvaluate_SyntaxError0() {
	_, err := Evaluate(")1>2", 0)
	t.EqualError(err, "parse error: syntax error: unexpected tokRPAREN, expecting tokIDENTIFIER or tokNUMBER or tokLPAREN\n)1>2\n^\n")
}

func (t *TestSuite) TestEvaluate_SyntaxError1() {
	_, err := Evaluate("01>2", 0)
	t.EqualError(err, "parse error: syntax error: unexpected tokNUMBER\n01>2\n ^\n")
}

func (t *TestSuite) TestEvaluate_SyntaxError2() {
	_, err := Evaluate("1>>2", 0)
	t.EqualError(err, "parse error: syntax error: unexpected tokGT, expecting tokIDENTIFIER or tokNUMBER or tokLPAREN\n1>>2\n  ^\n")
}

func (t *TestSuite) TestYYLex_num_Invalid() {
	lex := yyLex{
		peek: 'a',
		line: []byte("a"),
	}
	lval := yySymType{}
	tok := lex.num('a', &lval)
	t.Equal(eof, tok)
	t.EqualError(lex.Err, "ERROR: strconv.ParseUint: parsing \"a\": invalid syntax. Bad number \"a\": strconv.ParseUint: parsing \"a\": invalid syntax")
}

func BenchmarkEvaluate(b *testing.B) {
	for idx, expression := range testExpressions {
		if idx > 4 {
			break
		}
		Evaluate(expression.Expression, expression.N)
	}
}
