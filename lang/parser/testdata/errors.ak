let a = 1; // ### error[semicolon]: ";" is not allowed
---
let n = 007 // ### numbers cannot have leading zeros
---
a += 1 // ### compound assignment += is not supported
---
let s = 'a' // ### single quotes are not allowed
---
let x = a < b < c // ### comparisons do not chain
---
let x = a in b not in c // ### in does not chain
---
let x = a[1:2] // ### slices are not supported in v1
---
let x = {a: 1, a: 2} // ### duplicate key "a"
---
let x = {a} // ### expected :, got "}"
---
let s = "${a + 1}" // ### only names, members and indexes are allowed inside ${ }
---
(a) = 1 // ### cannot assign to this expression
---
(a).b = 1 // ### cannot assign to this expression
---
_ = 1 // ### cannot assign to this expression
---
x // ### unused value
---
let x // ### let x requires an initializer
---
var _ = 1 // ### _ can only be used with let
---
let if = 1 // ### "if" is a keyword and cannot be used as a name
---
let If = 1 // ### "If" is a keyword and cannot be used as a name
---
let fn = 1 // ### "fn" is reserved for a later version
---
if a {} // ### expected a new line after {
---
if a {
}
else { // ### else must be on the same line as the closing }
}
---
if {a: 1} { // ### an object literal cannot start a header
}
---
if a { // ### missing } for the block opened on line 4
---
break // ### break outside of a loop
---
if a {
    Continue // ### continue outside of a loop
}
---
for let x in a { // ### for let is redundant
}
---
for i, v range [0..3] { // ### range takes one loop variable
}
---
for i range 0..3 { // ### expected [ after range
}
---
let a = 1 let b = 2 // ### expected end of line
---
let x = 1 = 2 // ### expected end of line, got "="
---
x = y = 1 // ### expected end of line, got "="
---
let x = a b // ### expected end of line, got identifier b
---
let x = a ! in b // ### 4:11: error[expected-newline]: expected end of line, got "!"
---
return 1, 2 // ### expected end of line, got ","
---
var x, y // ### expected end of line, got ","
---
let = 1 // ### expected a name, got "="
---
let 1 = 2 // ### expected a name, got number 1
---
var fn // ### "fn" is reserved for a later version
---
catch(a) // ### "catch" is reserved for a later version
---
let x = try // ### expected an expression, got reserved try
---
var x = // ### expected an expression, got end of file
---
let x = not // ### expected an expression, got end of file
---
let x = 1 +
2 // ### 4:12: error[expected-expr]: expected an expression, got end of line
---
let x = a
.b // ### 5:1: error[expected-expr]: expected an expression, got "."
---
let x = a < b > c // ### comparisons do not chain
---
let x = a == b != c // ### comparisons do not chain
---
let x = a not in b in c // ### in does not chain
---
let x = (a // ### expected ), got end of file
---
let x = [1, 2 // ### expected ], got end of file
---
let x = f(1 2) // ### expected ), got number 2
---
let x = [,] // ### expected an expression, got ","
---
let x = f(,) // ### expected an expression, got ","
---
let x = f(a,, b) // ### expected an expression, got ","
---
let x = a[] // ### expected an expression, got "]"
---
let x = a[1:] // ### slices are not supported in v1
---
let x = a.if // ### expected a member name after ., got "if"
---
let x = a.1 // ### expected a member name after ., got number 1
---
let x = {1: 2} // ### invalid object key number 1
---
let x = {if: 1} // ### invalid object key "if"
---
let x = {a: 1,, b: 2} // ### invalid object key ","
---
let x = {"${a}": 1} // ### computed keys are not supported in v1
---
let x = {a: 1, "a": 2} // ### duplicate key "a"
---
let x = {a 1} // ### expected :, got number 1
---
let x = {a: } // ### expected an expression, got "}"
---
let x = {a: 1 b: 2} // ### expected }, got identifier b
---
let n = 007 // ### 4:9: error[leading-zero]
---
let s = "a\qb" // ### 4:11: error[invalid-escape]: unknown escape \q
---
let x = a @ b // ### 4:11: error[unexpected-char]
---
let s = "${a b}" // ### 4:14: error[interp-expr]: unexpected identifier b inside ${ }
---
let s = "${f()}" // ### 4:12: error[interp-expr]: only names, members and indexes are allowed inside ${ }
---
let s = "${a[b + 1]}" // ### only names, members and indexes are allowed inside ${ }
---
let s = "${1}" // ### only names, members and indexes are allowed inside ${ }
---
let s = "${a +}" // ### 4:15: error[expected-expr]: expected an expression, got end of file
---
let s = "${a["\q"]}" // ### 4:15: error[invalid-escape]
---
f().x = 1 // ### cannot assign to this expression
---
a + b = 1 // ### cannot assign to this expression
---
-a = 1 // ### cannot assign to this expression
---
_.x = 1 // ### cannot assign to this expression
---
a == 1 // ### 4:1: error[unused-value]
---
"x" // ### unused value
---
{a: 1} // ### unused value
---
a.b // ### unused value
---
if a
{ // ### 4:5: error[block-open]: expected { on the header line
}
---
if { // ### expected a condition before {
}
---
if a {
} else x // ### expected {, got identifier x
---
if a {
} else // ### expected {, got end of file
---
if a {
} else
{ // ### 5:7: error[block-open]
}
---
if a {
} else { f() // ### expected a new line after {, got identifier f
}
---
if a {
    f() g() // ### expected end of line, got identifier g
}
---
if a {
} else {
} else { // ### expected end of line, got "else"
}
---
if a {
} let x = 1 // ### expected end of line, got "let"
---
if a {
}
} // ### unexpected }
---
for x in a {
    if x { // ### missing } for the block opened on line 4
}
---
for x in { // ### expected a condition before {
}
---
for x in {a: 1} { // ### an object literal cannot start a header
}
---
for x { // ### expected in or range, got "{"
}
---
for in a { // ### "in" is a keyword and cannot be used as a name
}
---
for x, in a { // ### "in" is a keyword and cannot be used as a name
}
---
for var x, y range [0..1] { // ### range takes one loop variable
}
---
for i range [0 1] { // ### expected .. in range, got number 1
}
---
for i range [0..3 { // ### expected ], got "{"
}
---
for i range [0...3] { // ### expected an expression, got "."
}
---
for x in a {
}
break // ### break outside of a loop
---
for x in a {
    if x {
        break
    }
}
continue // ### continue outside of a loop
