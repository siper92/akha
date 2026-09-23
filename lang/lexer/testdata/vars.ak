// variables
Ak.Allow(FS...)
let name = "world"
name = "akha"
let content = FS.ReadFile("f.txt")
let n = 1 + 2 * 3
let s = "a" + "b"
let neg = -(n - 10) / 2 % 3
let ok = n > 3 and name != "" or not (s == "ab")
let le = n <= 7
let ge = n >= 7
let lt = n < 7
Ak.Log(s)
Ak.Exit("done", code=n)
