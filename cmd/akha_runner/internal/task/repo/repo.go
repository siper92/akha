package repo

import (
	"embed"
	"io/fs"
)

//go:embed all:data
var data embed.FS

func GetTaskFlow() (fs.FS, error) {
	task, err := fs.Sub(data, "data/task_proto")
	if err != nil {
		return nil, err
	}

	return task, nil
}
