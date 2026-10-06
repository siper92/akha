package parser

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

var SampleCRLF = "let a = 1\r\n\r\nif a {\r\n\n\n\n\n\n\n\n\n\t\t\t\n    let b = \"x\"\r\n}\r\n"

var SampleCRLFCanonical = `let a = 1
if a {
    let b = "x"
}`
