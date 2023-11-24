package main

import "testing"

func Test(t *testing.T) {
	testCases := []struct {
		desc                 string
		inputRuleFilePath    string
		inputDirPathGoModule string
	}{
		{
			desc:                 "",
			inputRuleFilePath:    "./examples/rulefiles/a.yaml",
			inputDirPathGoModule: "./examples/go_projects/prj01",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

		})
	}
}
