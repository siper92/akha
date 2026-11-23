package v3

const DefaultModel = "claude-opus-5-5"

type settings struct {
	cacheDir string
	workDir  string
}

var Config = settings{cacheDir: ".cache", workDir: "out"}

func (s settings) GetCacheDir() string { return s.cacheDir }
func (s settings) GetWorkDir() string  { return s.workDir }
