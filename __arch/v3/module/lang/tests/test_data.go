package tests

var Sample = `// a small step
let name = "ak"
var n = 0

if n < 3 and not false {
    n = n + 1
} else {
    return
}

for var i, v in [1, 2] {
    continue
}

return {name: name, total: n}
`

var SampleCanonical = `let name = "ak"
var n = 0
if n < 3 and not false {
    n = n + 1
} else {
    return
}
for var i, v in [1, 2] {
    continue
}
return {name: name, total: n}`

var SampleCRLF = "let a = 1\r\n\r\nif a {\r\n    let b = \"x\"\r\n}\r\n"

var SampleCRLFCanonical = `let a = 1
if a {
    let b = "x"
}`

var SpecDef = `// akha v1 - the spec by example
// one statement per line, ` + "`//`" + ` comments run to the end of the line
// indentation has no meaning, blocks use { on the header line and } on its own line

// --- let: immutable bindings, an initializer is required
let dir = "hello/" // comment statement
let count = 3.0                 // number, float64, prints as 3
let ratio = 0.5
let neg = -3.2                  // unary minus on a literal
let i = 0
let bool1 = true
let nothing = null
let arr = [1, 2, 3.2]           // array, mixed kinds allowed
let arr2 = ["a", "b", "c"]
let arr3 = ["1", 2, true, 3.2, [1, {"1": 1, "2": 2}]]
let obj = {"1": true, "2": 2, "3": 3.2}
let user = {name: "ann", "content-type": "json", tags: []}
let empty = {}

// --- multi line literals, newlines are ignored inside ( [ and an object {
let matrix = [
    [1, 2],
    [3, 4],                     // trailing comma allowed
]
let config = {
    retries: 3,
    endpoint: {host: "localhost", port: 8080},
}
let long = (1 +
    2 +
    3)

// --- var: mutable bindings
var total = 0
var last                        // starts as null
total = total + 1               // plain assignment, no += -= *= /= %=
last = "done"                   // null can take any kind

// --- discard: evaluates the expression, drops the value
let _ = count * 2

// --- strings: escapes and interpolation
let quoted = "say \"hi\"\n\ttab \\ back"
let greeting = "hello ${user.name}, first is ${arr[0]}"
let literal = "costs \${price}" // escaped, not interpolated

// --- operators, lowest to highest: or, and, not, in, compare, + -, * / %, unary -, postfix
let sum = 1 + 2 * 3             // 7
let grouped = (1 + 2) * 3       // 9
let rest = 10 % 3               // 1
let half = 7 / 2                // 3.5, float division
let joined = "ab" + "cd"        // string + string
let merged = arr + arr2         // array + array, a new array
let cmp = count >= 3 and not (i == 1)
let either = false or bool1
let has = 2 in arr              // any element == value
let missing = "x" not in obj    // key does not exist
let sub = "ell" in "hello"      // substring
let alias = true && 1 < 2 || !bool1

// --- access
let first = arr[0]
let name = user.name            // member access on objects
let ctype = user["content-type"]
let deep = arr3[4][1]["2"]
let key = "name"
let byKey = user[key]           // the index must be a string for objects

// --- input: the predeclared, immutable output of the previous step
let items = input.items

// --- nested assignment on var values, values are copied on assignment
var state = {count: 0, items: [{name: "a"}]}
state.count = state.count + 1
state["count"] = 2
state.items[0].name = "b"

// --- if / else if / else, conditions use truthiness
if true {
    // do something
}

// multi condition if statement
if bool1 and 1 < 2 and count > 2 {
    // do something
}

if count > 10 {
    total = 10
} else if count > 2 {
    total = count
} else {
    total = 0
}

// an object literal cannot start a header, wrap it in ( )
if ("name" in {name: 1}) {
}

// --- for in: arrays and objects
for var i in arr3 {             // shadows the outer i, var makes it mutable
    if i == 2 {
        break
    } else if i == "1" {
        continue
    } else {
        i = null                // only the loop variable changes
    }
}

for idx, item in arr {          // index, element, immutable
    total = total + idx
}

for k, v in obj {               // key, value in insertion order
    let pair = "${k}=${v}"
}

for _, v in arr2 {              // _ discards the index
    last = v
}

// --- for range: [start..end], end excluded
for var i range [0..10] {
    // do something with i
}

for n range [0..count] {
    total = total + n
}

// --- blocks open a new scope, a nested let or var shadows the parent binding
if total > 0 {
    let total = "shadowed"      // the outer total is untouched after the block
    var inner = 1
    inner = inner + 1
}

// --- step output: return ends the script, exit is an alias
return {total: total, greeting: greeting}
`
