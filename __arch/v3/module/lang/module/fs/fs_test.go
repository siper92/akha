package fs_test

import (
	"errors"
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
	"github.com/siper92/akha/lang/tests_utils/pipeline_test"
)

func TestFSWrite(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- write appends, overwrite replaces
		{
			Name: "write_and_overwrite",
			Input: `ak.import("fs")
fs.write("n.txt", "a")
fs.write("n.txt", "b")
fs.overwrite("m.txt", "a")
fs.overwrite("m.txt", "b")
return [fs.read("n.txt"), fs.read("m.txt"), fs.exists("x.txt")]`,
			Expected: `["ab","b",false]`,
		},
		// --- bigger files are built line by line in nested dirs
		{
			Name: "append_big_file",
			Input: `ak.import("fs")
var expected = ""
for i range [0..200] {
    fs.write("logs/day/big.txt", "line ${i}\n")
    expected = expected + "line ${i}\n"
}
return fs.read("logs/day/big.txt") == expected`,
			Expected: "true",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline_test.ValidateRunFS(t, c, "testdata")
		})
	}
}

func TestFSPath(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- separators and leading slashes resolve inside the root
		{
			Name: "normalized_path",
			Input: `ak.import("fs")
fs.overwrite("a\\b\\c.txt", "x")
return fs.read("/a/b/./c.txt")`,
			Expected: `"x"`,
		},
		// --- paths outside the root fail at runtime
		{
			Name:  "escape_root",
			Input: "ak.import(\"fs\")\nreturn fs.read(\"a/../../secret.txt\")",
			Err:   errors.New("invalid path: escapes root"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline_test.ValidateRunFS(t, c, "testdata")
		})
	}
}

func TestFSJSON(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- nested json data is read into objects and arrays
		{
			Name: "read_nested",
			Input: `ak.import("fs")
let d = fs.readJSON("users.json")
return [d.users[1].name, d.users[0].address.city, d.limits.cpu, d.users[1].manager]`,
			Expected: `["bob","Sofia",2.5,null]`,
		},
		// --- written json is indented, keeps key order and reads back equal
		{
			Name: "write_round_trip",
			Input: `ak.import("fs")
let d = fs.readJSON("users.json")
fs.writeJSON("out/users.json", d)
fs.writeJSON("small.json", {b: 1, a: [true, null]})
return [fs.readJSON("out/users.json") == d, fs.read("small.json")]`,
			Expected: `[true,"{\n  \"b\": 1,\n  \"a\": [\n    true,\n    null\n  ]\n}\n"]`,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline_test.ValidateRunFS(t, c, "testdata")
		})
	}
}

func TestFSYAML(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- anchors, flow lists and block strings are resolved
		{
			Name: "read_nested",
			Input: `ak.import("fs")
let c = fs.readYAML("config.yaml")
return [c.version, c.steps[0].opts.retries, c.steps[0].tags[1], c.steps[1].enabled, c.steps[1].script]`,
			Expected: `[2,3,"io",false,"line one\nline two\n"]`,
		},
		// --- yaml and json carry the same data
		{
			Name: "cross_format_round_trip",
			Input: `ak.import("fs")
let c = fs.readYAML("config.yaml")
fs.writeYAML("out/config.yaml", c)
fs.writeJSON("out/config.json", c)
return fs.readYAML("out/config.yaml") == fs.readJSON("out/config.json")`,
			Expected: "true",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline_test.ValidateRunFS(t, c, "testdata")
		})
	}
}
