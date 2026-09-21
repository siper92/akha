// utils - Ak always available
Ak.Allow(FS...) - allows the use of a module/namespace
Ak.Setup(
    log="info.log",
    debug="test.log", // don't need "," but it is allowed
) - sets up logging and debug output files

Ak.Log(message) - logs a message to - file
Ak.Debug(message, ...args) - logs a debug message to - file/cache directory
Ak.Exit("error", code=0) - exits with an error message and code, code 0 is default

// work with files
FS.ReadFile("file.txt") - gets content
FS.WriteFile("file.txt", "content") - writes content to file
FS.UpdateFile("file.txt", "new content") - updates file with new content
// work with directories
FS.ListFiles("path/") - lists all files in a directory
