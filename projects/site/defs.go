package site

import (
	"embed"
	_ "embed"
)

//go:embed _defs/db/schema.sql
var Schema string

//go:embed _defs/flow.yaml
var FlowDef string

//go:embed all:_defs/steps
var Steps embed.FS
