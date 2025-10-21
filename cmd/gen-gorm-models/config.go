package main

import (
	"gorm.io/gen/field"
	"log"
	"path/filepath"
	"runtime"

	"gorm.io/gen"
)

type Query interface {
	Filter(name string) ([]gen.T, error)
}

type GenerateConfig struct {
	ModelPackagePath string
	GenerateFunc     func(g *gen.Generator)
}

var generateConfig = GenerateConfig{"internal/system-watch/model", GenerateSystemWatchModels}

func createGenerator(packagePath string) *gen.Generator {
	return gen.NewGenerator(gen.Config{
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,
		ModelPkgPath:      packagePath,
		WithUnitTest:      true,
		FieldNullable:     true,
		FieldSignable:     false,
		FieldWithIndexTag: false,
		FieldWithTypeTag:  false,
	})
}

func applyGeneratorOptions(g *gen.Generator) {
	g.WithOpts(
		gen.FieldGORMTag("createdAt", func(tag field.GormTag) field.GormTag {
			tag.Set("default", "current_timestamp")
			return tag
		}),
		gen.FieldGORMTag("updatedAt", func(tag field.GormTag) field.GormTag {
			tag.Set("default", "current_timestamp")
			return tag
		}),
	)
}

func GenerateSystemWatchModels(g *gen.Generator) {
	g.GenerateModelAs("task", "TaskM")
}

func rootDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("Error retrieving file info")
	}

	dir := filepath.Dir(file)

	absPath, err := filepath.Abs(dir + "../../../")
	if err != nil {
		log.Fatalf("Error getting absolute directory path: %v", dir)
	}

	return absPath
}

func resolveModelPackagePath(defaultPath string) string {
	if *modelPath != "" {
		return *modelPath
	}

	return filepath.Join(rootDir(), defaultPath)
}
