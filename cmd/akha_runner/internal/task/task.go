package task

import (
	"io/fs"

	"github.com/siper92/akha/cmd/akha_runner/internal/task/repo"
)

type Step interface {
	Name() string
	Content() string
}

var _ Step = (*TaskStep)(nil)

type TaskStep struct {
	name    string
	content string
}

func (t TaskStep) Name() string {
	return t.name
}

func (t TaskStep) Content() string {
	return t.content
}

func GetTaskSteps() ([]Step, error) {
	taskProto, err := repo.GetTaskFlow()
	if err != nil {
		return nil, err
	}

	files := []string{}
	err = fs.WalkDir(
		taskProto, ".",
		func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				files = append(files, p)
			}

			return nil
		})
	if err != nil {
		return nil, err
	}

	tasks := []Step{}
	for _, file := range files {
		content, err := fs.ReadFile(taskProto, file)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, &TaskStep{
			name:    file,
			content: string(content),
		})
	}

	return tasks, nil
}
