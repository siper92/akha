Ak.Allow(FS...)
Ak.Setup(log="hello.log", debug="hello.debug.log")

let dir = "hello/"
let count = 3
let i = 0

while i < count {
    i = i + 1
    if i == 2 {
        continue
    }

    FS.WriteFile(dir + "file" + Ak.Str(i) + ".txt", "hello " + Ak.Str(i))
}

let files = FS.ListFiles(dir)
for f in files {
    let content = FS.ReadFile(dir + f)
    Ak.Debug("file", f, content)
    if Ak.Len(content) > 6 {
        Ak.Log("long: " + f)
    } else {
        Ak.Log("short: " + f)
    }
}

if Ak.Len(files) == 2 {
    Ak.Exit("done", code=0)
} else {
    Ak.Exit("unexpected file count " + Ak.Str(Ak.Len(files)), code=1)
}
