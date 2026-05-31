package ast

import (
	"personal-assistant-server/global"
	"path/filepath"
	"testing"
)

func TestPackageInitializeGorm_Injection(t *testing.T) {
	type fields struct {
		Type        Type
		Path        string
		ImportPath  string
		StructName  string
		PackageName string
		IsNew       bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			fields: fields{
				Type:        TypePackageInitializeGorm,
				Path:        filepath.Join(global.GVA_CONFIG.AutoCode.Root, global.GVA_CONFIG.AutoCode.Server, "initialize", "gorm_biz.go"),
				StructName:  "ExaFileUploadAndDownload",
				IsNew:       false,
			},
		},
		{
			fields: fields{
				Type:        TypePackageInitializeGorm,
				Path:        filepath.Join(global.GVA_CONFIG.AutoCode.Root, global.GVA_CONFIG.AutoCode.Server, "initialize", "gorm_biz.go"),
				StructName:  "ExaCustomer",
				IsNew:       false,
			},
		},
		{
			fields: fields{
				Type:        TypePackageInitializeGorm,
				Path:        filepath.Join(global.GVA_CONFIG.AutoCode.Root, global.GVA_CONFIG.AutoCode.Server, "initialize", "gorm_biz.go"),
				StructName:  "ExaFileUploadAndDownload",
				IsNew:       true,
			},
		},
		{
			fields: fields{
				Type:        TypePackageInitializeGorm,
				Path:        filepath.Join(global.GVA_CONFIG.AutoCode.Root, global.GVA_CONFIG.AutoCode.Server, "initialize", "gorm_biz.go"),
				StructName:  "ExaCustomer",
				IsNew:       true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &PackageInitializeGorm{
				Type:        tt.fields.Type,
				Path:        tt.fields.Path,
				ImportPath:  tt.fields.ImportPath,
				StructName:  tt.fields.StructName,
				PackageName: tt.fields.PackageName,
				IsNew:       tt.fields.IsNew,
			}
			file, err := a.Parse(a.Path, nil)
			if err != nil {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			a.Injection(file)
			err = a.Format(a.Path, nil, file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Injection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPackageInitializeGorm_Rollback(t *testing.T) {
	type fields struct {
		Type        Type
		Path        string
		ImportPath  string
		StructName  string
		PackageName string
		IsNew       bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			fields: fields{
				Type:        TypePackageInitializeGorm,
				Path:        filepath.Join(global.GVA_CONFIG.AutoCode.Root, global.GVA_CONFIG.AutoCode.Server, "initialize", "gorm_biz.go"),
				StructName:  "ExaFileUploadAndDownload",
				IsNew:       false,
			},
		},
		{
			fields: fields{
				Type:        TypePackageInitializeGorm,
				Path:        filepath.Join(global.GVA_CONFIG.AutoCode.Root, global.GVA_CONFIG.AutoCode.Server, "initialize", "gorm_biz.go"),
				StructName:  "ExaCustomer",
				IsNew:       false,
			},
		},
		{
			fields: fields{
				Type:        TypePackageInitializeGorm,
				Path:        filepath.Join(global.GVA_CONFIG.AutoCode.Root, global.GVA_CONFIG.AutoCode.Server, "initialize", "gorm_biz.go"),
				StructName:  "ExaFileUploadAndDownload",
				IsNew:       true,
			},
		},
		{
			fields: fields{
				Type:        TypePackageInitializeGorm,
				Path:        filepath.Join(global.GVA_CONFIG.AutoCode.Root, global.GVA_CONFIG.AutoCode.Server, "initialize", "gorm_biz.go"),
				StructName:  "ExaCustomer",
				IsNew:       true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &PackageInitializeGorm{
				Type:        tt.fields.Type,
				Path:        tt.fields.Path,
				ImportPath:  tt.fields.ImportPath,
				StructName:  tt.fields.StructName,
				PackageName: tt.fields.PackageName,
				IsNew:       tt.fields.IsNew,
			}
			file, err := a.Parse(a.Path, nil)
			if err != nil {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			a.Rollback(file)
			err = a.Format(a.Path, nil, file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
