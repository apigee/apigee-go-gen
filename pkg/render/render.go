// Copyright 2025 Google LLC
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
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/apigee/apigee-go-gen/pkg/flags"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-errors/errors"
	"github.com/gosimple/slug"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"text/template"
)

func RenderGeneric(context any, cFlags *CommonFlags, dryRun bool) error {
	var err error

	outputDir := filepath.Dir(string(cFlags.OutputFile))
	//create the output directory
	if !dryRun {
		err = os.MkdirAll(outputDir, os.ModePerm)
		if err != nil {
			return errors.New(err)
		}
	}

	//create the template
	tmpl, err := CreateTemplate(string(cFlags.TemplateFile), string(cFlags.TemplateFileAlias), cFlags.IncludeList, string(cFlags.OutputFile), dryRun)
	if err != nil {
		return err
	}

	//render the template
	rendered, err := RenderTemplate(tmpl, context)
	if err != nil {
		return err
	}

	//write rendered template to output
	useStdout := false
	if cFlags.OutputFile == "-" || dryRun == true {
		useStdout = true
	}

	if useStdout {
		fmt.Print(string(rendered))
	} else {
		err = os.WriteFile(string(cFlags.OutputFile), rendered, os.ModePerm)
		if err != nil {
			return errors.New(err)
		}
	}

	return nil
}

func RenderTemplate(tmpl *template.Template, context any) ([]byte, error) {
	renderedBytes := bytes.Buffer{}
	err := tmpl.Execute(&renderedBytes, context)
	if err != nil {
		return nil, errors.New(err)
	}

	//remove empty lines with just #
	regex := regexp.MustCompile(`(?ms)^\s*#\s*$[\r\n]*`)
	replaced := regex.ReplaceAll(renderedBytes.Bytes(), []byte{})

	//remove lines with just # // a comment
	regex = regexp.MustCompile(`(?m)^\s*#\s*//.+$[\r\n]*`)
	replaced = regex.ReplaceAll(replaced, []byte{})
	return replaced, nil
}

func CreateTemplate(templateFile string, templateFileAlias string, includeList []string, outputFile string, dryRun bool) (*template.Template, error) {
	var err error
	var includeMatches []string

	if _, err = os.Stat(templateFile); err != nil {
		return nil, errors.New(err)
	}

	// add helper files if present
	templateDir := filepath.Dir(templateFile)

	helpersFileNames := []string{"_helpers.tpl", "_helpers.tmpl"}
	for _, helperFileName := range helpersFileNames {
		helperFilePath := filepath.Join(templateDir, helperFileName)
		if _, err = os.Stat(helperFilePath); err == nil {
			includeList = append(includeList, helperFilePath)
		}
	}

	//add the main template itself
	includeList = append(includeList, templateFile)

	if includeMatches, err = ExpandInclude(includeList); err != nil {
		return nil, err
	}

	helperFuncs := map[string]any{}

	blankFunc := func(args ...any) string {
		return ""
	}

	slugMakeFunc := func(args ...any) string {
		defer recoverPanic()
		if len(args) == 0 {
			panic("slug_make function requires one argument")
		}
		in := args[0].(string)
		return slug.Make(in)
	}

	derefFunc := func(args ...any) any {
		defer recoverPanic()
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
		defer recoverPanic()
		if len(args) == 0 {
			panic("os_getenv function requires one argument")
		}

		envName := args[0].(string)
		return os.Getenv(envName)
	}

	osGetEnvs := func(args ...any) map[string]string {
		defer recoverPanic()
		envs := map[string]string{}
		for _, envPair := range os.Environ() {
			key, value, _ := strings.Cut(envPair, "=")
			envs[key] = value
		}

		return envs
	}

	urlParseFunc := func(args ...any) url.URL {
		defer recoverPanic()
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

	osWritefileFunc := func(args ...any) string {
		defer recoverPanic()
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
		defer recoverPanic()
		if len(args) < 0 {
			panic("fmt_printf function requires at least one argument")
		}

		fmt.Printf(args[0].(string), args[1:]...)
		return ""
	}

	includeStack := []string{fmt.Sprintf("file:%s", templateFile)}

	includeFunc := func(args ...any) string {
		defer recoverPanic()
		if len(args) < 0 {
			panic("include function requires at least one argument")
		}

		arg0 := args[0].(string)

		var err error
		var templateBytes []byte
		var templateName string

		parentTemplateIndex := slices.IndexFunc(includeStack, func(elem string) bool {
			return strings.Index(elem, "file:") == 0
		})

		parentTemplateFile := strings.TrimPrefix(includeStack[parentTemplateIndex], "file:")
		targetTemplateFile := filepath.Join(filepath.Dir(parentTemplateFile), arg0)

		if templateBytes, err = os.ReadFile(targetTemplateFile); err == nil {
			// file was found, use its contents as template
			templateName = arg0
			includeStack = slices.Insert(includeStack, 0, fmt.Sprintf("file:%s", targetTemplateFile))

		} else {
			//no such file, assume it's a named template
			templateBytes = []byte(fmt.Sprintf(`{{- template "%s" . }}`, arg0))
			templateName = fmt.Sprintf("%s.tpl", arg0)
			includeStack = slices.Insert(includeStack, 0, fmt.Sprintf("template:%s", templateName))
		}

		templateText := string(templateBytes)

		tpl, err := template.New(templateName).
			Funcs(helperFuncs).
			Funcs(sprig.FuncMap()).
			Parse(templateText)
		if err != nil {
			panic(err)
		}

		if len(includeMatches) > 0 {
			tpl, err = tpl.ParseFiles(includeMatches...)
			if err != nil {
				panic(err)
			}
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

		includeStack = slices.Delete(includeStack, 0, 1)

		return tplOut.String()
	}

	helperFuncs["include"] = includeFunc
	helperFuncs["fmt_printf"] = fmtPrintfFunc
	helperFuncs["os_writefile"] = osWritefileFunc
	helperFuncs["url_parse"] = urlParseFunc
	helperFuncs["os_getenv"] = osGetEnvFunc
	helperFuncs["os_getenvs"] = osGetEnvs
	helperFuncs["os_copyfile"] = getOSCopyFileFunc(templateFile, outputFile, dryRun)
	helperFuncs["remove_oas_extensions"] = getRemoveOASExtensions(templateFile, outputFile, dryRun)
	helperFuncs["blank"] = blankFunc
	helperFuncs["deref"] = derefFunc
	helperFuncs["slug_make"] = slugMakeFunc
	helperFuncs["oas3_to_mcp"] = convertOAS3ToMCPValues
	helperFuncs["yaml_to_json"] = convertYAMLTextToJSON
	helperFuncs["json_to_yaml"] = convertJSONToYAML

	var templateText []byte
	if templateText, err = os.ReadFile(templateFile); err != nil {
		return nil, errors.New(err)
	}

	if templateFileAlias == "" {
		templateFileAlias = templateFile
	}

	tmpl, err := template.New(templateFileAlias).
		Funcs(helperFuncs).
		Funcs(sprig.FuncMap()).
		Parse(string(templateText))
	if err != nil {
		return nil, errors.New(err)
	}

	if len(includeMatches) > 0 {
		tmpl, err = tmpl.ParseFiles(includeMatches...)
		if err != nil {
			return nil, errors.New(err)
		}
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
