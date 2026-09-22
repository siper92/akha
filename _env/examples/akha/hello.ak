// hello.ak - basic integration script
Ak.Allow(FS...)
Ak.Setup(
    log="hello.log",
    debug="hello.debug.log",
)

Ak.Log("hello from akha")
Ak.Debug("debug args", 1, "two", true)

FS.WriteFile("hello/hello.txt", "hello")
FS.UpdateFile("hello/hello.txt", "hello, world")
FS.ReadFile("hello/hello.txt")
FS.ListFiles("hello/")

Ak.Exit("done", code=0)
