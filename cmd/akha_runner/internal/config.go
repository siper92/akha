package internal

type Config struct {
	RunDir string `yaml:"run_dir"`
	DbName string `yaml:"db_name"`
	AI     struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"ai"`
}
