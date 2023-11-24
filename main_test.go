package main

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	testCases := []struct {
		desc                 string
		inputRuleFilePath    string
		inputDirPathGoModule string
		expectedOutput       string
	}{
		{
			desc:                 "",
			inputRuleFilePath:    "./examples/rulefiles/rule1.yaml",
			inputDirPathGoModule: "./examples/go_projects/prj01",
			expectedOutput: `## www.example.com/hoge/fuga/pkg1

下記ファイルに違反があります。

- examples/go_projects/prj01/pkg1/pkg1.go
  - "import www.example.com/hoge/fuga/pkg2"はルール"ルール１"に違反します。
  - "import www.example.com/hoge/fuga/pkg3"はルール"ルール１"に違反します。

## www.example.com/hoge/fuga/pkg1/pkg11

下記ファイルに違反があります。

- examples/go_projects/prj01/pkg1/pkg11/pkg11.go
  - "import www.example.com/hoge/fuga/pkg2"はルール"ルール１"に違反します。
  - "import www.example.com/hoge/fuga/pkg3"はルール"ルール１"に違反します。

`,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			output := bytes.NewBufferString("")
			log.SetFlags(0)
			log.SetOutput(output)
			run(tC.inputRuleFilePath, tC.inputDirPathGoModule)
			assert.Equal(t, tC.expectedOutput, output.String())
		})
	}
}
