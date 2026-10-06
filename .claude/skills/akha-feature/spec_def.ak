// akha v1 - the spec by example, a flow step that reports on active users
// one statement per line, `//` comments run to the end of the line
// blocks use { on the header line and } on its own line, indentation has no meaning

// --- modules: ak is always imported, others are imported at the root block
ak.import("fs")
ak.setup(log="info")                // named args only
ak.log("report step started")

// --- let is deep immutable, var is mutable, var alone starts as null
let outDir = "out/"
let minAge = 18                     // one number kind, float64, 18.0 prints as 18
let offset = -0.5                   // unary minus on a literal
var total = 0
var last
let _ = minAge * offset             // _ evaluates and drops the value
let previous = input                // output of the previous step, null when none

// --- module calls return values, fs paths stay inside the worker root
let data = fs.readJSON("users.json")
let config = fs.readYAML("config.yaml")
let users = data.users
let retries = config.defaults.retries
let city = users[0]["address"].city // index and member access, a missing key is null

// --- literals span lines inside ( [ {, a trailing comma is allowed
var report = {
    team: data.team,
    "content-type": "json",
    active: [],
    roles: {},
    total: 0,
}

// --- for in: arrays give index, element, objects give key, value in insertion order
for var user in users {             // var makes the loop variable mutable
    if not user.active or user.age < minAge {
        continue
    } else if user.name == "stop" {
        break                       // leaves the innermost loop
    }
    report.active = report.active + [user.name]
    for _, role in user.roles {     // _ discards the index
        if role not in report.roles {
            report.roles[role] = 0
        }
        report.roles[role] = report.roles[role] + 1
    }
    user = null                     // only the loop variable changes
}

// --- for range: [start..end], end excluded, bounds evaluated once
for n range [0..retries] {
    fs.write("${outDir}run.log", "attempt ${n}\n")  // write appends
    total = total + n               // plain assignment, no += operators
}

// --- operators, lowest to highest: or, and, not, in, compare, + -, * / %, unary -, postfix
let ratio = total / 2               // float division, 1.5
let ready = total >= 3 && !(ratio == 1)
let admin = "admin" in report.roles or "x" in ["a", "b"]
let greeting = "team ${report.team} in ${city}, total ${total}"
let literal = "costs \${price}"     // escaped, not interpolated

// --- blocks open a scope, a nested let shadows the parent binding
if ready {
    let total = "shadowed"
    last = total                    // null can take any kind
}

// an object literal cannot start a header, wrap it in ( )
if ("team" in {team: 1}) {
    ak.debug(greeting)
} else {
    ak.debug(literal)
}

// --- files: overwrite replaces, json and yaml keep key order
report.total = total
fs.overwrite("${outDir}summary.txt", "${greeting}\n")
fs.writeJSON("${outDir}report.json", report)
fs.writeYAML("${outDir}report.yaml", report)
let saved = fs.readJSON("${outDir}report.json") == report

// --- step output: return ends the script, exit is an alias
return {
    team: report.team,
    active: report.active,
    roles: report.roles,
    total: total,
    ready: ready,
    admin: admin,
    last: last,
    saved: saved,
    previous: previous,
}
