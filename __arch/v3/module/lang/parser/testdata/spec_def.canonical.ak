let dir = "hello/"
let count = 3
let ratio = 0.5
let neg = -3.2
let i = 0
let bool1 = true
let nothing = null
let arr = [1, 2, 3.2]
let arr2 = ["a", "b", "c"]
let arr3 = ["1", 2, true, 3.2, [1, {"1": 1, "2": 2}]]
let obj = {"1": true, "2": 2, "3": 3.2}
let user = {name: "ann", "content-type": "json", tags: []}
let empty = {}
let matrix = [[1, 2], [3, 4]]
let config = {retries: 3, endpoint: {host: "localhost", port: 8080}}
let long = 1 + 2 + 3
var total = 0
var last
total = total + 1
last = "done"
let _ = count * 2
let quoted = "say \"hi\"\n\ttab \\ back"
let greeting = "hello ${user.name}, first is ${arr[0]}"
let literal = "costs \${price}"
let sum = 1 + 2 * 3
let grouped = (1 + 2) * 3
let rest = 10 % 3
let half = 7 / 2
let joined = "ab" + "cd"
let merged = arr + arr2
let cmp = count >= 3 and not i == 1
let either = false or bool1
let has = 2 in arr
let missing = "x" not in obj
let sub = "ell" in "hello"
let alias = true and 1 < 2 or not bool1
let first = arr[0]
let name = user.name
let ctype = user["content-type"]
let deep = arr3[4][1]["2"]
let key = "name"
let byKey = user[key]
let items = input.items
var state = {count: 0, items: [{name: "a"}]}
state.count = state.count + 1
state["count"] = 2
state.items[0].name = "b"
if true {
}
if bool1 and 1 < 2 and count > 2 {
}
if count > 10 {
    total = 10
} else if count > 2 {
    total = count
} else {
    total = 0
}
if "name" in {name: 1} {
}
for var i in arr3 {
    if i == 2 {
        break
    } else if i == "1" {
        continue
    } else {
        i = null
    }
}
for idx, item in arr {
    total = total + idx
}
for k, v in obj {
    let pair = "${k}=${v}"
}
for _, v in arr2 {
    last = v
}
for var i range [0..10] {
}
for n range [0..count] {
    total = total + n
}
if total > 0 {
    let total = "shadowed"
    var inner = 1
    inner = inner + 1
}
return {total: total, greeting: greeting}