package std

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/siper92/akha/lang/module"
	"github.com/siper92/akha/lang/module/ak"
	"github.com/siper92/akha/lang/module/fs"
)

var Default = sync.OnceValue(func() module.IRegistry {
	reg, err := New(".", slog.Default())
	if err != nil {
		panic(err)
	}
	return reg
})

func New(root string, logger *slog.Logger) (module.IRegistry, error) {
	core, errCore := ak.New(logger)
	files, errFiles := fs.New(root)
	if err := errors.Join(errCore, errFiles); err != nil {
		return nil, err
	}

	return module.NewRegistry(core, files)
}
