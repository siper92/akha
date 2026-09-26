package lexer_test

import (
	"strings"
	"testing"

	"github.com/siper92/akha/lang/tests"
)

func TestLexOK(t *testing.T) {
	// --- ident, keyword, reserved
	tests.RunLexOK(t, []tests.LexOKCase{
		{
			Name: "all_keywords",
			Src:  strings.Join(tests.LexKeywords, " "),
		},
		{
			Name: "reserved_words",
			Src:  strings.Join(tests.LexReserved, " "),
		},
		{
			Name: "case_sensitive_names",
			Src:  "True False Null Let IF",
		},
		{
			Name: "underscore_names",
			Src:  "_ _a _1 a_b A_B_9",
		},
	})

	// --- number
	tests.RunLexOK(t, []tests.LexOKCase{
		{
			Name: "integers",
			Src:  "0 42 8080 10",
		},
		{
			Name: "floats",
			Src:  "3.2 1.32332 1.0 0.5 0.0",
		},
		{
			Name: "negative_literals",
			Src:  "-3 -3.2 -0.5 -0.1 -0",
		},
		{
			Name: "range_bounds",
			Src:  "[0..10]",
		},
	})

	// --- string, escape, interp
	tests.RunLexOK(t, []tests.LexOKCase{
		{
			Name: "empty_string",
			Src:  `""`,
		},
		{
			Name: "plain_string",
			Src:  `"hello/"`,
		},
		{
			Name: "all_escapes",
			Src:  `"\n \t \r \" \\ \$"`,
		},
		{
			Name: "spec_escapes",
			Src:  `"say \"hi\"\n\ttab \\ back"`,
		},
		{
			Name: "interp_ident",
			Src:  `"hello ${name}"`,
		},
		{
			Name: "interp_member_and_index",
			Src:  `"hello ${user.name}, first is ${arr[0]}"`,
		},
		{
			Name: "interp_string_index",
			Src:  `"${user["content-type"]}"`,
		},
		{
			Name: "interp_nested_value_ref",
			Src:  `"${a[b].c[0]}"`,
		},
		{
			Name: "interp_adjacent",
			Src:  `"${k}=${v}"`,
		},
		{
			Name: "escaped_interp",
			Src:  `"costs \${price}"`,
		},
		{
			Name: "dollar_without_brace",
			Src:  `"cost $5 and $"`,
		},
		{
			Name: "braces_without_dollar",
			Src:  `"a { b } c"`,
		},
		{
			Name: "comment_marker_in_string",
			Src:  `"http://x // y"`,
		},
		{
			Name: "semicolon_and_quote_in_string",
			Src:  `"a; 'b' & c | d"`,
		},
		{
			Name: "unicode_in_string",
			Src:  `"héllo 日本語"`,
		},
	})

	// --- comment
	tests.RunLexOK(t, []tests.LexOKCase{
		{
			Name: "line_comment",
			Src:  "// a comment",
		},
		{
			Name: "empty_comment",
			Src:  "//",
		},
		{
			Name: "comment_after_statement",
			Src:  "let a = 1 // trailing",
		},
		{
			Name: "comment_hides_bad_runes",
			Src:  "// ; ' & | += @ # 007 \\q",
		},
		{
			Name: "unicode_in_comment",
			Src:  "// héllo 日本語",
		},
		{
			Name: "comment_before_eof_without_newline",
			Src:  "let a = 1\n// end",
		},
	})

	// --- operator, punct
	tests.RunLexOK(t, []tests.LexOKCase{
		{
			Name: "arithmetic",
			Src:  "+ - * / %",
		},
		{
			Name: "comparison",
			Src:  "== != < <= > >=",
		},
		{
			Name: "logical_aliases",
			Src:  "&& || !",
		},
		{
			Name: "assign",
			Src:  "a = 1",
		},
		{
			Name: "punct",
			Src:  ", : . .. ( ) [ ] { }",
		},
	})

	// --- source, space, NL
	tests.RunLexOK(t, []tests.LexOKCase{
		{
			Name: "empty_source",
			Src:  "",
		},
		{
			Name: "only_newlines",
			Src:  "\n\n\n",
		},
		{
			Name: "crlf_lines",
			Src:  "let a = 1\r\nlet b = 2\r\n",
		},
		{
			Name: "bom_prefix",
			Src:  "let a = 1\n",
		},
		{
			Name: "tabs_and_cr_are_space",
			Src:  "\tlet\ta\r=\t1",
		},
		{
			Name: "no_trailing_newline",
			Src:  "let a = 1",
		},
	})

	// --- samples
	tests.RunLexOK(t, []tests.LexOKCase{
		{
			Name: "sample",
			Src:  tests.Sample,
		},
		{
			Name: "sample_crlf",
			Src:  tests.SampleCRLF,
		},
		{
			Name: "spec_def",
			Src:  tests.SpecDef,
		},
	})
}

func TestLexTokenCount(t *testing.T) {
	// --- numbers_are_single_tokens
	tests.RunLexCount(t, []tests.LexSameCase{
		{
			Name: "zero",
			Src:  "0",
			Ref:  "x",
		},
		{
			Name: "integer",
			Src:  "42",
			Ref:  "x",
		},
		{
			Name: "float",
			Src:  "3.2",
			Ref:  "x",
		},
		{
			Name: "float_zero_fraction",
			Src:  "1.0",
			Ref:  "x",
		},
		{
			Name: "negative_is_minus_and_number",
			Src:  "-3",
			Ref:  "a b",
		},
		{
			Name: "negative_float_is_minus_and_number",
			Src:  "-3.2",
			Ref:  "a b",
		},
		{
			Name: "range_is_number_dotdot_number",
			Src:  "0..10",
			Ref:  "a b c",
		},
		{
			Name: "range_in_brackets",
			Src:  "[0..count]",
			Ref:  "a b c d e",
		},
		{
			Name: "float_then_dotdot",
			Src:  "1.5..2",
			Ref:  "a b c",
		},
	})

	// --- strings_are_single_tokens
	tests.RunLexCount(t, []tests.LexSameCase{
		{
			Name: "plain",
			Src:  `"a b c"`,
			Ref:  "x",
		},
		{
			Name: "escaped_quote",
			Src:  `"a \" b"`,
			Ref:  "x",
		},
		{
			Name: "comment_marker_inside",
			Src:  `"a // b"`,
			Ref:  "x",
		},
		{
			Name: "escaped_interp",
			Src:  `"\${x}"`,
			Ref:  "x",
		},
		{
			Name: "dollar_alone",
			Src:  `"$x"`,
			Ref:  "x",
		},
	})

	// --- operators_longest_match
	tests.RunLexCount(t, []tests.LexSameCase{
		{
			Name: "eq",
			Src:  "a == b",
			Ref:  "a b c",
		},
		{
			Name: "neq",
			Src:  "a != b",
			Ref:  "a b c",
		},
		{
			Name: "le",
			Src:  "a <= b",
			Ref:  "a b c",
		},
		{
			Name: "ge",
			Src:  "a >= b",
			Ref:  "a b c",
		},
		{
			Name: "and_alias",
			Src:  "a && b",
			Ref:  "a b c",
		},
		{
			Name: "or_alias",
			Src:  "a || b",
			Ref:  "a b c",
		},
		{
			Name: "not_alias",
			Src:  "!a",
			Ref:  "a b",
		},
		{
			Name: "no_spaces",
			Src:  "a<b",
			Ref:  "a b c",
		},
		{
			Name: "member",
			Src:  "a.b",
			Ref:  "a b c",
		},
		{
			Name: "dotdot_idents",
			Src:  "a..b",
			Ref:  "a b c",
		},
		{
			Name: "eq_then_not",
			Src:  "a==!b",
			Ref:  "a b c d",
		},
		{
			Name: "assign_then_minus",
			Src:  "a=-1",
			Ref:  "a b c d",
		},
	})

	// --- spaces_and_comments_are_skipped
	tests.RunLexCount(t, []tests.LexSameCase{
		{
			Name: "indentation",
			Src:  "    let   a =   1",
			Ref:  "let a = 1",
		},
		{
			Name: "tabs",
			Src:  "a\t\tb",
			Ref:  "a b",
		},
		{
			Name: "cr_is_space",
			Src:  "a\rb",
			Ref:  "a b",
		},
		{
			Name: "trailing_comment",
			Src:  "let a = 1 // c d e",
			Ref:  "let a = 1",
		},
		{
			Name: "comment_no_space",
			Src:  "a//b",
			Ref:  "a",
		},
	})
}

func TestLexNormalize(t *testing.T) {
	// --- crlf_is_newline
	tests.RunLexSame(t, []tests.LexSameCase{
		{
			Name: "crlf_lines",
			Src:  "let a = 1\r\nlet b = 2\r\n",
			Ref:  "let a = 1\nlet b = 2\n",
		},
		{
			Name: "crlf_blank_lines",
			Src:  "a\r\n\r\nb",
			Ref:  "a\n\nb",
		},
		{
			Name: "crlf_after_comment",
			Src:  "a // c\r\nb",
			Ref:  "a // c\nb",
		},
		{
			Name: "sample_crlf",
			Src:  tests.SampleCRLF,
			Ref:  strings.ReplaceAll(tests.SampleCRLF, "\r\n", "\n"),
		},
	})

	// --- bom_is_ignored
	tests.RunLexSame(t, []tests.LexSameCase{
		{
			Name: "bom_statement",
			Src:  "let a = 1",
			Ref:  "let a = 1",
		},
		{
			Name: "bom_multi_line",
			Src:  "a\nb",
			Ref:  "a\nb",
		},
	})

	// --- deterministic
	tests.RunLexSame(t, []tests.LexSameCase{
		{
			Name: "spec_def_twice",
			Src:  tests.SpecDef,
			Ref:  tests.SpecDef,
		},
	})
}

func TestLexErr(t *testing.T) {
	// --- semicolon
	tests.RunLexErr(t, []tests.LexErrCase{
		{
			Name: "semicolon_alone",
			Src:  ";",
			Line: 1,
			Col:  1,
		},
		{
			Name: "semicolon_after_statement",
			Src:  "let a = 1;",
			Line: 1,
			Col:  10,
		},
		{
			Name: "semicolon_between_statements",
			Src:  "let a = 1; let b = 2",
			Line: 1,
			Col:  10,
		},
	})

	// --- strings
	tests.RunLexErr(t, []tests.LexErrCase{
		{
			Name: "single_quote",
			Src:  "let s = 'a'",
			Line: 1,
			Col:  9,
		},
		{
			Name: "unknown_escape",
			Src:  `let s = "a\q"`,
			Line: 1,
		},
		{
			Name: "unknown_escape_unicode",
			Src:  `let s = "A"`,
			Line: 1,
		},
		{
			Name: "unterminated_string",
			Src:  `let s = "abc`,
			Line: 1,
		},
		{
			Name: "unterminated_after_escaped_quote",
			Src:  `let s = "abc\"`,
			Line: 1,
		},
		{
			Name: "raw_newline_in_string",
			Src:  "let s = \"ab\ncd\"",
			Line: 1,
		},
		{
			Name: "raw_crlf_in_string",
			Src:  "let s = \"ab\r\ncd\"",
			Line: 1,
		},
		{
			Name: "unterminated_interp",
			Src:  `let s = "${name"`,
			Line: 1,
		},
	})

	// --- numbers
	tests.RunLexErr(t, []tests.LexErrCase{
		{
			Name: "leading_zeros",
			Src:  "let n = 007",
			Line: 1,
			Col:  9,
		},
		{
			Name: "double_zero",
			Src:  "let n = 00",
			Line: 1,
			Col:  9,
		},
		{
			Name: "leading_zero_float",
			Src:  "let n = 01.5",
			Line: 1,
			Col:  9,
		},
		{
			Name: "followed_by_letter",
			Src:  "let n = 3abc",
			Line: 1,
		},
		{
			Name: "followed_by_underscore",
			Src:  "let n = 3_000",
			Line: 1,
		},
		{
			Name: "exponent",
			Src:  "let n = 1e3",
			Line: 1,
		},
		{
			Name: "float_exponent",
			Src:  "let n = 1.5e3",
			Line: 1,
		},
		{
			Name: "hex",
			Src:  "let n = 0x1F",
			Line: 1,
		},
		{
			Name: "octal_prefix",
			Src:  "let n = 0o17",
			Line: 1,
		},
	})

	// --- operators
	tests.RunLexErr(t, []tests.LexErrCase{
		{
			Name: "single_ampersand",
			Src:  "a & b",
			Line: 1,
			Col:  3,
		},
		{
			Name: "single_pipe",
			Src:  "a | b",
			Line: 1,
			Col:  3,
		},
		{
			Name: "plus_assign",
			Src:  "a += 1",
			Line: 1,
			Col:  3,
		},
		{
			Name: "minus_assign",
			Src:  "a -= 1",
			Line: 1,
			Col:  3,
		},
		{
			Name: "star_assign",
			Src:  "a *= 1",
			Line: 1,
			Col:  3,
		},
		{
			Name: "slash_assign",
			Src:  "a /= 1",
			Line: 1,
			Col:  3,
		},
		{
			Name: "percent_assign",
			Src:  "a %= 1",
			Line: 1,
			Col:  3,
		},
	})

	// --- unknown_runes
	tests.RunLexErr(t, []tests.LexErrCase{
		{
			Name: "at_sign",
			Src:  "a @ b",
			Line: 1,
			Col:  3,
		},
		{
			Name: "hash",
			Src:  "# not a comment",
			Line: 1,
			Col:  1,
		},
		{
			Name: "dollar_outside_string",
			Src:  "let a = $b",
			Line: 1,
			Col:  9,
		},
		{
			Name: "backtick",
			Src:  "let a = `b`",
			Line: 1,
			Col:  9,
		},
		{
			Name: "question_mark",
			Src:  "a ? b",
			Line: 1,
			Col:  3,
		},
		{
			Name: "non_ascii_ident_start",
			Src:  "é = 1",
			Line: 1,
			Col:  1,
		},
		{
			Name: "non_ascii_ident_rest",
			Src:  "café = 1",
			Line: 1,
			Col:  4,
		},
	})

	// --- encoding
	tests.RunLexErr(t, []tests.LexErrCase{
		{
			Name: "invalid_utf8_alone",
			Src:  "\xff",
			Line: 1,
			Col:  1,
		},
		{
			Name: "invalid_utf8_in_string",
			Src:  "let a = \"\xff\"",
			Line: 1,
			Col:  10,
		},
		{
			Name: "invalid_utf8_in_comment",
			Src:  "a\n// \xfe",
			Line: 2,
			Col:  4,
		},
		{
			Name: "bom_not_counted",
			Src:  ";",
			Line: 1,
			Col:  1,
		},
		{
			Name: "crlf_is_one_line",
			Src:  "a\r\n\r\n;",
			Line: 3,
			Col:  1,
		},
	})

	// --- positions_count_runes
	tests.RunLexErr(t, []tests.LexErrCase{
		{
			Name: "tab_is_one_column",
			Src:  "\tlet a;",
			Line: 1,
			Col:  7,
		},
		{
			Name: "accented_rune_in_string",
			Src:  `let s = "héllo" ;`,
			Line: 1,
			Col:  17,
		},
		{
			Name: "cjk_runes_in_string",
			Src:  `"日本語" ;`,
			Line: 1,
			Col:  7,
		},
		{
			Name: "runes_in_comment_then_next_line",
			Src:  "// 日本語\n  ;",
			Line: 2,
			Col:  3,
		},
		{
			Name: "escape_is_not_a_newline",
			Src:  `"a\nb" ;`,
			Line: 1,
			Col:  8,
		},
		{
			Name: "third_line_indented",
			Src:  "let a = 1\n\n    let b = 2 ;",
			Line: 3,
			Col:  15,
		},
	})

	// --- first_error_over_whole_source
	tests.RunLexErr(t, []tests.LexErrCase{
		{
			Name: "first_of_two_errors",
			Src:  "let a = 1;\nlet b = 'x'",
			Line: 1,
			Col:  10,
		},
		{
			Name: "first_of_two_on_one_line",
			Src:  "a & b | c",
			Line: 1,
			Col:  3,
		},
		{
			Name: "parse_errors_do_not_stop_lexing",
			Src:  "let let let\n) ) (\n;",
			Line: 3,
			Col:  1,
		},
		{
			Name: "reserved_words_lex_then_error",
			Src:  "fn try catch ;",
			Line: 1,
			Col:  14,
		},
		{
			Name: "comment_content_is_skipped",
			Src:  "// ; ' & |\n;",
			Line: 2,
			Col:  1,
		},
		{
			Name: "error_after_spec_def",
			Src:  tests.SpecDef + "let z = 1;\n",
			Line: 141,
			Col:  10,
		},
	})
}

func TestLexSemicolonHint(t *testing.T) {
	// --- semicolon_hint
	tests.RunLexHint(t, []tests.LexOKCase{
		{
			Name: "semicolon_alone",
			Src:  ";",
		},
		{
			Name: "semicolon_after_statement",
			Src:  "let a = 1;",
		},
	}, tests.LexHintSemicolon)
}
