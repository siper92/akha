package check_test

import (
	"errors"
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
	"github.com/siper92/akha/lang/tests_utils/pipeline"
)

func TestCheckScopes(t *testing.T) {
	cases := []tests_utils.Case[string, struct{}]{
		// --- names resolve from the current block to the root
		{
			Name:  "shadow_in_nested_block",
			Input: "var a = 1\nif a {\n    let a = \"x\"\n}\na = 2",
		},
		{
			Name:  "use_before_declaration",
			Input: "let b = a\nlet a = 1",
			Err:   errors.New("1: error[use-before-decl]: a is used before its declaration"),
		},
		{
			Name:  "redeclare_in_same_block",
			Input: "let a = 1\nvar a = 2",
			Err:   errors.New("2: error[redeclared]: a is already declared in this block"),
		},
		{
			Name:  "unreachable_after_exit",
			Input: "for x in input {\n    exit x\n    let y = 1\n}",
			Err:   errors.New("3: error[unreachable]: unreachable code"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateCheck(t, c)
		})
	}
}

func TestCheckAssign(t *testing.T) {
	cases := []tests_utils.Case[string, struct{}]{
		// --- deep targets resolve to the root binding
		{
			Name:  "deep_target_on_var",
			Input: "var s = {items: [{name: \"a\"}]}\ns.items[0].name = \"b\"",
		},
		{
			Name:  "deep_target_on_let",
			Input: "let s = {items: [1]}\ns.items[0] = 2",
			Err:   errors.New("2: error[immutable]: cannot assign to let s"),
		},
		{
			Name:  "immutable_loop_variable",
			Input: "for x in input {\n    x = 1\n}",
			Err:   errors.New("2: error[immutable]: cannot assign to loop variable x"),
		},
		{
			Name:  "later_declared_target",
			Input: "a = 1\nvar a",
			Err:   errors.New("1: error[use-before-decl]: a is used before its declaration"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateCheck(t, c)
		})
	}
}

func TestCheckErrors(t *testing.T) {
	cases := []tests_utils.Case[string, struct{}]{
		// --- static errors stop the script before eval
		{
			Name:  "undeclared_name",
			Input: "let a = 1\nlet b = c",
			Err:   errors.New("2: error[undeclared]: c is not declared"),
		},
		{
			Name:  "assign_to_input",
			Input: "input.name = \"x\"",
			Err:   errors.New("1: error[immutable]: cannot assign to input"),
		},
		{
			Name:  "unknown_callee",
			Input: "let a = send(1)",
			Err:   errors.New("1: error[no-callable]: unknown callee send, v1 has no callables"),
		},
		{
			Name:  "member_on_known_number",
			Input: "let n = 1\nlet x = n.size",
			Err:   errors.New("2: error[member-kind]: member access works only on objects"),
		},
		// --- known values and nested scopes fail at check time
		{
			Name:  "division_by_known_zero",
			Input: "let d = 0\nlet x = 1 / d",
			Err:   errors.New("2: error[div-zero]: division by zero"),
		},
		{
			Name:  "member_on_known_array",
			Input: "let a = [1]\nlet n = a.size",
			Err:   errors.New("2: error[member-kind]: member access works only on objects"),
		},
		{
			Name:  "assign_to_undeclared",
			Input: "x = 1",
			Err:   errors.New("1: error[undeclared]: x is not declared"),
		},
		{
			Name:  "nested_use_before_declaration",
			Input: "if input {\n    let b = a\n}\nlet a = 1",
			Err:   errors.New("2: error[use-before-decl]: a is used before its declaration"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateCheck(t, c)
		})
	}
}

func TestCheckKnownValues(t *testing.T) {
	cases := []tests_utils.Case[string, struct{}]{
		// --- let bindings keep their known value for the value checkers
		{
			Name:  "known_member_through_let",
			Input: "let cfg = {db: {host: \"h\"}}\nlet h = cfg.db.host",
		},
		{
			Name:  "missing_member_through_let",
			Input: "let cfg = {db: {host: \"h\"}}\nlet p = cfg.db.port",
			Err:   errors.New(`2: error[member-missing]: member "port" does not exist on the object`),
		},
		{
			Name:  "modulo_by_known_zero",
			Input: "let zero = 0\nlet x = 10 % zero",
			Err:   errors.New("2: error[div-zero]: modulo by zero"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateCheck(t, c)
		})
	}
}
