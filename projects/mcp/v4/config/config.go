package config

const (
	cacheDir      = ".cache"
	workDir       = "workspace"
	model         = "claude-opus-5-5"
	maxTokens     = 16000
	maxToolRounds = 20
)

func GetCacheDir() string { return cacheDir }

func GetWorkDir() string { return workDir }

func GetModel() string { return model }

func GetMaxTokens() int64 { return maxTokens }

func GetMaxToolRounds() int { return maxToolRounds }
