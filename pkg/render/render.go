// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package render

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-errors/errors"
	"github.com/gosimple/slug"
	"github.com/micovery/apigee-yaml-toolkit/pkg/flags"
	"github.com/micovery/apigee-yaml-toolkit/pkg/utils"
	"github.com/micovery/apigee-yaml-toolkit/pkg/values"
	"google.golang.org/protobuf/types/descriptorpb"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"text/template"
)

func RenderGeneric(context any, flags *Flags) error {
	var err error

	type RenderContext struct {
		Proto    descriptorpb.FileDescriptorProto
		ProtoStr string

		Values map[string]any
	}

	outputDir := filepath.Dir(flags.OutputFile)
	//create the output directory
	if !flags.DryRun {
		err = os.MkdirAll(outputDir, os.ModePerm)
		if err != nil {
			return errors.New(err)
		}
	}

	//create the template
	tmpl, err := CreateTemplate(flags.TemplateFile, flags.IncludeList, flags.OutputFile, flags.DryRun)
	if err != nil {
		return err
	}

	//render the template
	rendered, err := RenderTemplate(tmpl, context)
	if err != nil {
		return err
	}

	//write rendered template to output
	if !flags.DryRun {
		err = os.WriteFile(flags.OutputFile, rendered, os.ModePerm)
		if err != nil {
			return errors.New(err)
		}
	}

	if flags.DryRun {
		fmt.Fprintf(os.Stdout, string(rendered))
	}

	return nil
}

func RenderTemplate(tmpl *template.Template, context any) ([]byte, error) {
	renderedBytes := bytes.Buffer{}
	err := tmpl.Execute(&renderedBytes, context)
	if err != nil {
		return nil, errors.New(err)
	}

	regex := regexp.MustCompile(`(?ms)^\s*#\s*$[\r\n]*`)
	replaced := regex.ReplaceAll(renderedBytes.Bytes(), []byte{})
	return replaced, nil
}

func CreateTemplate(templateFile string, includeList []string, outputFile string, dryRun bool) (*template.Template, error) {
	includeMatches, err := ExpandInclude(includeList)
	if err != nil {
		return nil, err
	}

	//make a copy of the template to create a unique temporary main file
	tmpFile, err := os.CreateTemp("", "tpl_*_"+filepath.Base(templateFile))
	if err != nil {
		return nil, errors.New(err)
	}
	defer tmpFile.Close()

	err = utils.CopyFile(tmpFile.Name(), templateFile)
	if err != nil {
		return nil, err
	}

	includeMatches = append([]string{tmpFile.Name()}, includeMatches...)

	helperFuncs := map[string]any{}

	blankFunc := func(args ...any) string {
		return ""
	}

	slugMakeFunc := func(args ...any) string {
		if len(args) == 0 {
			panic("slug_make function requires one argument")
		}
		in := args[0].(string)
		return slug.Make(in)
	}

	derefFunc := func(args ...any) any {
		if len(args) == 0 {
			panic("deref function requires one argument")
		}
		//use reflection to deref the pointer
		valOf := reflect.ValueOf(args[0])
		if valOf.Kind() == reflect.Ptr {
			return valOf.Elem().Interface()
		} else {
			return args[0]
		}
	}

	osGetEnvFunc := func(args ...any) string {
		if len(args) == 0 {
			panic("os_getenv function requires one argument")
		}

		envName := args[0].(string)
		return os.Getenv(envName)
	}

	osGetEnvs := func(args ...any) map[string]string {
		envs := map[string]string{}
		for _, envPair := range os.Environ() {
			key, value, _ := strings.Cut(envPair, "=")
			envs[key] = value
		}

		return envs
	}

	urlParseFunc := func(args ...any) url.URL {
		if len(args) == 0 {
			panic("url_parse function requires one argument")
		}

		urlIn := args[0].(string)
		result, err := url.Parse(urlIn)
		if err != nil {
			panic(err)
		}
		return *result
	}

	osCopyFileFunc := func(args ...any) string {
		if len(args) < 2 {
			panic("os_copyfile function requires two arguments")
		}

		dst := args[0].(string)
		src := args[1].(string)

		if filepath.IsAbs(dst) {
			panic("os_copyfile dst must not be absolute")
		}
		if strings.Index(dst, "..") >= 0 {
			panic("os_copyfile dst must not use ..")
		}

		if filepath.IsAbs(src) {
			panic("os_copyfile src must not be absolute")
		}
		if strings.Index(src, "..") >= 0 {
			panic("os_copyfile src must not use ..")
		}

		dstDir := filepath.Dir(dst)
		if !dryRun {
			err := os.MkdirAll(dstDir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}

		srcPath := filepath.Join(filepath.Dir(templateFile), src)
		dstPath := filepath.Join(filepath.Dir(outputFile), dst)

		if !dryRun {
			srcFileContent, err := os.ReadFile(srcPath)
			if err != nil {
				panic(err)
			}

			err = os.WriteFile(dstPath, srcFileContent, os.ModePerm)
			if err != nil {
				panic(err)
			}
		} else {
			//fmt.Printf(`os_copyfile("%s", "%s")\n`, dstPath, srcPath)
		}

		return dst
	}

	osWritefileFunc := func(args ...any) string {
		if len(args) < 2 {
			panic("os_writefile function requires two arguments")
		}

		fileName := args[0].(string)
		if filepath.IsAbs(fileName) {
			panic("os_writefile path must not be absolute")
		}
		if strings.Index(fileName, "..") >= 0 {
			panic("os_writefile path must not use ..")
		}

		filePath := filepath.Join(filepath.Dir(outputFile), fileName)
		data := args[1].(string)
		var err error
		if !dryRun {
			err = os.WriteFile(filePath, []byte(data), os.ModePerm)
			if err != nil {
				panic(err)
			}
		} else {
			//fmt.Printf(`os_writefile(%s, []byte{%v})\n`, filePath, len(data))
		}
		return fileName
	}

	fmtPrintfFunc := func(args ...any) string {
		if len(args) < 0 {
			panic("fmt_printf function requires at least one argument")
		}

		fmt.Printf(args[0].(string), args[1:]...)
		return ""
	}

	includeFunc := func(args ...any) string {
		if len(args) < 0 {
			panic("include function requires at least one argument")
		}

		templateName := args[0].(string)
		templateText := fmt.Sprintf(`{{- template "%s" . }}`, templateName)

		tpl, _ := template.New(templateName + ".tpl").
			Funcs(helperFuncs).
			Funcs(sprig.FuncMap()).
			Parse(templateText)
		tpl, err := tpl.ParseFiles(includeMatches...)
		if err != nil {
			panic(err)
		}

		var arg any
		if len(args) > 1 {
			//actual argument
			arg = args[1]
		} else {
			//empty arguments
			arg = map[string]any{}
		}
		tplOut := bytes.Buffer{}
		err = tpl.Execute(&tplOut, arg)
		if err != nil {
			panic(err)
		}
		return tplOut.String()
	}

	helperFuncs["include"] = includeFunc
	helperFuncs["fmt_printf"] = fmtPrintfFunc
	helperFuncs["os_writefile"] = osWritefileFunc
	helperFuncs["url_parse"] = urlParseFunc
	helperFuncs["os_getenv"] = osGetEnvFunc
	helperFuncs["os_getenvs"] = osGetEnvs
	helperFuncs["os_copyfile"] = osCopyFileFunc
	helperFuncs["blank"] = blankFunc
	helperFuncs["deref"] = derefFunc
	helperFuncs["slug_make"] = slugMakeFunc

	tmpl, err := template.New(filepath.Base(tmpFile.Name())).
		Funcs(helperFuncs).
		Funcs(sprig.FuncMap()).
		ParseFiles(includeMatches...)
	if err != nil {
		return nil, errors.New(err)
	}

	return tmpl, nil
}

func ExpandInclude(includeTpl flags.IncludeList) ([]string, error) {
	// expand the included templates
	allMatches := []string{}
	for _, includePattern := range includeTpl {
		basePath, pattern := doublestar.SplitPattern(includePattern)
		fSys := os.DirFS(basePath)
		matches, err := doublestar.Glob(fSys, pattern)
		if err != nil {
			return nil, errors.New(err)
		}

		if len(matches) == 0 {
			return nil, errors.Errorf(`include pattern "%s" did not match any file`, includePattern)
		}

		for _, match := range matches {
			newMatch := filepath.Join(basePath, match)
			allMatches = append(allMatches, newMatch)
		}

	}

	return allMatches, nil
}

type Flags struct {
	Version      bool
	TemplateFile string
	OutputFile   string
	DryRun       bool
	Help         bool
	IncludeList  flags.IncludeList
	Values       *values.Map
}

func GetRenderFlags(printVersion func(), printUsage func()) (*Flags, error) {

	rFlags := &Flags{
		Values: &values.Map{},
	}

	dryRun := flags.Bool(false)
	setValue := flags.NewSetAny(rFlags.Values)
	setValueStr := flags.NewSetString(rFlags.Values)
	setValueFile := flags.NewValues(rFlags.Values)

	flag.BoolVar(&rFlags.Version, "version", false, "(optional) prints version text")
	flag.BoolVar(&rFlags.Help, "help", false, `(optional) prints additional help text`)
	flag.StringVar(&rFlags.TemplateFile, "template", "", `(required) path to main template e.g. "./input.tpl"`)
	flag.Var(&rFlags.IncludeList, "include", `(optional,0..*) path/glob for templates to include e.g. "./helpers/*.tpl"`)
	flag.StringVar(&rFlags.OutputFile, "output", "", "(required) output directory")
	flag.Var(&setValue, "set", `(optional,0..*) sets a context key/value  .e.g "boolVal=true" or "floatVal=1.37"`)
	flag.Var(&setValueStr, "set-string", `(optional,0..*) sets a context key/value .e.g key1="value1"`)
	flag.Var(&dryRun, "dry-run", "(optional) prints the rendered template")
	flag.Var(&setValueFile, "values", "(optional,0..1) sets context keys/values from YAML file")

	flag.Parse()

	rFlags.DryRun = bool(dryRun) //this hack is needed so user does not need to use equals sign i.e. --dry-run=false

	if rFlags.Version {
		printVersion()
		os.Exit(0)
		return nil, nil
	}

	if rFlags.Help {
		printUsage()
		os.Exit(0)
		return nil, nil
	}

	if strings.TrimSpace(rFlags.TemplateFile) == "" {
		utils.RequireParamAndExit("template")
		return nil, nil
	}

	if strings.TrimSpace(rFlags.OutputFile) == "" && !rFlags.DryRun {
		utils.RequireParamAndExit("output")
		return nil, nil
	}

	return rFlags, nil
}
