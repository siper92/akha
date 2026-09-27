// chunks are split by a line with only ---
// an expected error is a trailing comment: // ### "message part"
// a chunk without a marker must parse
let a = 1
var b
b = a + 1
---
let a = 1; // ### "is not allowed"
---
let n = 007 // ### "numbers cannot have leading zeros"
---
a += 1 // ### "compound assignment += is not supported"
---
let s = 'a' // ### "single quotes are not allowed"
---
let x = a < b < c // ### "comparisons do not chain"
---
let x = a in b not in c // ### "in does not chain"
---
let x = a[1:2] // ### "slices are not supported in v1"
---
let x = {a: 1, a: 2} // ### "duplicate key \"a\""
---
let x = {a} // ### "expected :, got \"}\""
---
let s = "${a + 1}" // ### "only names, members and indexes are allowed inside ${ }"
---
(a) = 1 // ### "cannot assign to this expression"
---
(a).b = 1 // ### "cannot assign to this expression"
---
_ = 1 // ### "cannot assign to this expression"
---
x // ### "unused value"
---
let x // ### "let x requires an initializer"
---
var _ = 1 // ### "_ can only be used with let"
---
let if = 1 // ### "\"if\" is a keyword and cannot be used as a name"
---
let If = 1 // ### "\"If\" is a keyword and cannot be used as a name"
---
let fn = 1 // ### "\"fn\" is reserved for a later version"
---
LET a = 1
IF a {
} ELSE IF a Not In [1] {
} Else {
}
---
if a {} // ### "expected a new line after {"
---
if a {
}
else { // ### "else must be on the same line as the closing }"
}
---
if {a: 1} { // ### "an object literal cannot start a header"
}
---
if a { // ### "missing } for the block opened on line 1"
---
break // ### "break outside of a loop"
---
if a {
    Continue // ### "continue outside of a loop"
}
---
for let x in a { // ### "for let is redundant"
}
---
for i, v range [0..3] { // ### "range takes one loop variable"
}
---
for i range 0..3 { // ### "expected [ after range"
}
---
for x in a {
    if x {
        break
    }
}
---
let a = 1 let b = 2 // ### "expected end of line"
