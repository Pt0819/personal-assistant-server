//go:generate go mod tidy
//go:generate go mod download
//go:generate go run gen.go
package main

import (
	"gorm.io/gen"
	"path/filepath"

	"personal-assistant-server/plugin/announcement/model"
)

func main() {
	g := gen.NewGenerator(gen.Config{OutPath: filepath.Join("..", "..", "..", "announcement", "blender", "model", "dao"), Mode: gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface})
	g.ApplyBasic(
		new(model.Info),
	)
	g.Execute()
}
