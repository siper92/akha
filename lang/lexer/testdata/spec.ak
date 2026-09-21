// utils - Ak always available
Ak.Allow(FS...)
Ak.Setup(
    log="info.log",
    debug="test.log", // trailing comma allowed
)
Ak.Log("hello")
Ak.Debug("dbg", 1, "two")
// work with files
FS.WriteFile("file.txt", "content")
FS.ReadFile("file.txt")
FS.UpdateFile("file.txt", "new content")
// work with directories
FS.ListFiles("path/")
Ak.Exit("done", code=0)
