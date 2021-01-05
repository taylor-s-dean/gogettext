package po2json

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	poFilePath = "../testdata/test.po"
)

type TestSuite struct {
	suite.Suite
}

func TestPO2JSON(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (t *TestSuite) TestLoadFile() {
	loader := Loader{}
	poJSON, err := loader.LoadFile(poFilePath)
	t.NoError(err)

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	t.NoError(enc.Encode(poJSON))
}

func BenchmarkLoadBytes(b *testing.B) {
	fileContents, err := ioutil.ReadFile(poFilePath)
	if err != nil {
		b.FailNow()
	}

	loader := Loader{}
	b.ResetTimer()

	for i := 0; i < 100; i++ {
		loader.LoadBytes(fileContents)
	}
}
