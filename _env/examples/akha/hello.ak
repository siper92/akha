// hello.ak - basic integration script
Ak.Allow(FS...)
Ak.Setup(
    log="worker.log",
    debug="worker.debug.log",
)

Ak.Log("akha log test - interpret")
Ak.Debug("debug with args", 1, "two", true)

FS.WriteFile("hello/hello.txt", "hello")
FS.UpdateFile("hello/hello.txt", "hello, world")
FS.ReadFile("hello/hello.txt")
FS.ListFiles("hello/")

Ak.Exit("done", code=0)
