// utils - Ak always available
Ak.Allow(FS...) - allows the use of a module/namespace, must be the first statement
Ak.Setup(
    log="info.log",
    debug="test.log", // don't need "," but it is allowed
) - sets up logging and debug output files

Ak.Log(message) - logs a message to - file
Ak.Debug(message, ...args) - logs a debug message to - file/cache directory
Ak.Exit("error", code=0) - exits with an error message and code, code 0 is default
Ak.Str(value) - string form of any value
Ak.Len(value) - length of a string or a list
Ak.Range(n) - list of ints from 0 to n-1

// work with files
FS.ReadFile("file.txt") - gets content
FS.WriteFile("file.txt", "content") - writes content to file
FS.UpdateFile("file.txt", "new content") - updates file with new content
// work with directories
FS.ListFiles("path/") - lists all files in a directory

// variables
let name = "world"              - declaration, error if name is already declared in the same scope
name = "akha"                   - assignment, error if name is not declared
let content = FS.ReadFile("f")  - a call is an expression, its result is the value
let n = 1 + 2 * 3               - int arithmetic
let s = "a" + "b"               - string concat
let ok = n > 3 and name != ""   - bool logic

// expressions
literals     "str" 42 true false
reference    name
call         Module.Func(args, key=value)
unary        -x  not ok
binary       + - * / %  == != < <= > >=  and or
grouping     (a + b) * c
precedence   or < and < not < comparison < + - < * / % < unary < primary

// conditionals
if n > 5 {
    Ak.Log("big")
} else if n > 2 {
    Ak.Log("mid")
} else {
    Ak.Log("small")
}

// loops
for f in FS.ListFiles("dir/") {
    Ak.Log(f)
}
let i = 0
while i < 3 {
    i = i + 1
    if i == 2 {
        continue
    }
    Ak.Log(Ak.Str(i))
}
break    - leaves the closest loop
continue - jumps to the next iteration of the closest loop

// rules
- a script is a list of statements, one per line
- statements: call, let, assign, if, for, while, break, continue
- only a call can stand alone as a statement
- a block is `{` on the header line, statements one per line, `}` on its own line
- `} else {` and `} else if cond {` stay on one line
- newlines are allowed only inside call parentheses
- every block opens a scope, `let` declares in the current scope, assignment finds the closest declaration
- `for` declares the loop variable in the loop scope
- conditions must be bool, anything else is a runtime error
- `for` iterates a list, anything else is a runtime error
- `+` works on int+int and string+string, `- * / %` on ints, `/` and `%` by zero is a runtime error
- `< <= > >=` compare ints or strings, `== !=` compare any values, different types are never equal
- `break` and `continue` outside a loop, undefined names and redeclarations are static check errors
- `Ak.Allow` outside the first statement or inside a block is a static check error

// grammar
script   = { stmt NEWLINE }
stmt     = call | "let" IDENT "=" expr | IDENT "=" expr | if | for | while | "break" | "continue"
if       = "if" expr block [ "else" ( if | block ) ]
for      = "for" IDENT "in" expr block
while    = "while" expr block
block    = "{" NEWLINE { stmt NEWLINE } "}"
call     = IDENT "." IDENT "(" [ arg { "," arg } [ "," ] ] ")"
arg      = IDENT "..." | IDENT "=" expr | expr
expr     = or
or       = and { "or" and }
and      = not { "and" not }
not      = "not" not | cmp
cmp      = add [ ( "==" | "!=" | "<" | "<=" | ">" | ">=" ) add ]
add      = mul { ( "+" | "-" ) mul }
mul      = unary { ( "*" | "/" | "%" ) unary }
unary    = "-" unary | primary
primary  = STRING | INT | "true" | "false" | call | IDENT | "(" expr ")"
