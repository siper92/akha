package fs_test

import (
	"os"
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
	"github.com/siper92/akha/lang/tests_utils/pipeline_test"
)

func TestSpecDef(t *testing.T) {
	src, err := os.ReadFile("../../../../.claude/skills/akha-feature/spec_def.ak")
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}

	cases := []tests_utils.Case[string, string]{
		// --- the skill example runs end to end on the fs testdata
		{
			Name:     "spec_def_report",
			Input:    string(src),
			Expected: `{"team":"core","active":["ann"],"roles":{"admin":1,"dev":1},"total":3,"ready":true,"admin":true,"last":"shadowed","saved":true,"previous":null}`,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline_test.ValidateRunFS(t, c, "testdata")
		})
	}
}
