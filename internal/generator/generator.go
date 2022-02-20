package generator

import (
	"bytes"
	"fmt"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func Generate(
	name string,
	userType string,
	userTypePackage string,
	workingDir string,
) error {
	data, err := getTemplateData(name, userType, userTypePackage, workingDir)
	if err != nil {
		return fmt.Errorf("cannot get template data: %w", err)
	}
	filename := fmt.Sprintf("%v_gen.go", strings.ToLower(name))
	path := filepath.Join(workingDir, filename)
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("cannot generate cache file: %w", err)
	}
	formattedBytes, err := imports.Process(path, buf.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("cannot gofmt generated cache file: %w", err)
	}
	if err = ioutil.WriteFile(path, formattedBytes, 0644); err != nil {
		return fmt.Errorf("cannot write generated file: %w", err)
	}
	return nil
}

type templateData struct {
	Name         string
	Package      string
	TypePackage  string
	FullTypeName string
}

func getTemplateData(
	name string,
	userType string,
	userTypePackage string,
	workingDir string,
) (templateData, error) {
	genPkg, err := getPackageName(workingDir)
	if err != nil {
		return templateData{}, err
	}
	data := templateData{
		Name:         name,
		Package:      genPkg.Name,
		FullTypeName: userType,
	}
	if userTypePackage != "" && genPkg.PkgPath != userTypePackage {
		data.FullTypeName = fmt.Sprintf("usertype.%v", userType)
		data.TypePackage = userTypePackage
	}
	return data, nil
}

func getPackageName(dir string) (*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{Dir: dir}, ".")
	if err != nil {
		return nil, getPackageNameErr(dir, err)
	}
	if len(pkgs) != 1 {
		return nil, getPackageNameErr(
			dir,
			fmt.Errorf("%v packages loaded, expected 1", len(pkgs)),
		)
	}
	return pkgs[0], nil
}

func getPackageNameErr(dir string, err error) error {
	return fmt.Errorf("cannot get package name for directory %v: %w", dir, err)
}
