package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

// VERSION ...
var VERSION string

// SourceFile ...
type SourceFile struct {
	name         string
	folder       string
	src          string
	contents     *string
	dependencies []SourceFile
}

var logger = loadLogger()

func loadLogger() *log.Logger {
	var _, debugFlag = os.LookupEnv("SCALA_SCRIPT_DEBUG")
	if debugFlag {
		return log.New(os.Stderr, "", 0)
	}

	null, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0666)
	var logger = log.New(null, "", 0)
	return logger
}

// cheap exit and error message
//
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// how to use
//
func usage() {
	sysLogger := log.New(os.Stderr, "", 0)
	name := filepath.Base(os.Args[0])
	sysLogger.Println(name, VERSION)
	sysLogger.Println("usage:", name, "script.scala [script-args]")
	sysLogger.Println("usage:", name, "--repl [scala-args]")
}

// one-liner text loader
//
func fileAsText(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	return string(data), err
}

// one-liner text storage
//
func textToFile(filename string, content string) error {
	bytes := []byte(content)
	err := ioutil.WriteFile(filename, bytes, 0640)
	return err
}

// recursively load included file contents
//
func loadChildFile(working string, filename string) (*SourceFile, error) {
	srcName, err := filepath.Abs(filepath.Join(working, filename))
	if err != nil {
		return nil, err
	}
	result := SourceFile{}
	result.name = filename
	result.folder = filepath.Dir(srcName)
	result.src = filepath.Base(srcName)

	// load the contents
	//
	contents, err := fileAsText(filepath.Join(result.folder, result.src))
	if err != nil {
		return nil, err
	}

	// scan through to get the list of files we need
	//
	re := regexp.MustCompile(`(?m)(//#include)(\W+)(.*)\n`)
	for _, match := range re.FindAllStringSubmatch(contents, -1) {
		childName := match[3]
		child, err := loadChildFile(result.folder, childName)
		if err != nil {
			return nil, err
		}
		result.dependencies = append(result.dependencies, *child)
	}

	// go through once more to inject the contents
	//
	for _, child := range result.dependencies {
		rer := regexp.MustCompile(`//#include\W+` + child.name)
		contents = rer.ReplaceAllLiteralString(contents, *child.contents)
	}

	result.contents = &contents

	return &result, nil
}

// root script file loader
//
func loadFile(working string, filename string) (*SourceFile, string, error) {

	result, err := loadChildFile(working, filename)
	if err != nil {
		return nil, "", err
	}
	destName := ".g." + result.src

	var changed bool
	outFile := filepath.Join(result.folder, destName)

	// get timestamp off the source file
	//
	ss, err := os.Stat(filepath.Join(result.folder, result.src))
	if err != nil {
		return nil, "", err
	}

	// dest file is optional
	//
	ds, err := os.Stat(outFile)
	if err != nil {
		// either it does not exist or some other reason we would want to overwrite it
		//
		changed = true
		logger.Println("no destination exists")
	} else if ds != nil {
		changed = ss.ModTime().After(ds.ModTime())
		if !changed {
			// final check is a content check against the generated content
			//
			otherContent, err := fileAsText(outFile)
			if err != nil {
				return nil, "", err
			}
			changed = otherContent != *result.contents
			if changed {
				logger.Println("content difference")
			}
		} else {
			logger.Println("timestamp difference")
		}
	}

	if changed {
		logger.Println("saving file")
		err = textToFile(outFile, *result.contents)
		check(err)
	}

	return result, outFile, nil
}

// class path builder
//
func gatherClassPath(root string) string {
	logger.Println("class path root:", root)
	files, err := ioutil.ReadDir(filepath.Join(root, "lib"))
	if err != nil {
		return ""
	}

	var cp []string
	for i := range files {
		f := files[i]
		if filepath.Ext(f.Name()) == ".jar" {
			name := filepath.Join(root, "lib", f.Name())
			cp = append(cp, name)
		}
	}
	cp = append(cp, root)

	return strings.Join(cp, ":")
}

// entry point
//
func main() {
	args := os.Args[1:]

	if len(args) <= 0 {
		usage()
		os.Exit(1)
	}

	wd, err := os.Getwd()
	check(err)

	var cp string
	var result *SourceFile
	var scriptFile string

	if args[0] == "--repl" {
		cp = gatherClassPath(wd)
		scriptFile = wd
	} else {
		result, scriptFile, err = loadFile(wd, args[0])
		check(err)

		cp = gatherClassPath(result.folder)
	}

	cmdFile, err := exec.LookPath("scala")
	check(err)

	env := os.Environ()
	launchArgs := []string{
		cmdFile,
		"-deprecation",
		"-feature",
		"-savecompiled",
		"-classpath",
		cp,
		"-Dscala.script.name=" + scriptFile}

	// first arg is either --repl or script to execute
	//
	if result != nil {
		launchArgs = append(launchArgs, scriptFile)
	}
	args = args[1:]
	launchArgs = append(launchArgs, args...)

	// spawn the call, replacing ourselves
	//
	logger.Println("command:", launchArgs)
	err = syscall.Exec(cmdFile, launchArgs, env)
	check(err)
}
