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
var sysLogger = loadSysLogger()

func loadSysLogger() *log.Logger {
	return log.New(os.Stderr, "", 0)
}

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
		return root
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

// Options ...
type Options struct {
	help       bool
	repl       bool
	nop        bool
	comp       bool
	scriptName string
	scalaArgs  []string
	scriptArgs []string
}

func parse(args []string) Options {
	if len(args) <= 0 {
		usage()
		os.Exit(0)
	}

	opts := Options{}
	opts.help = false
	opts.repl = false
	opts.nop = false
	opts.comp = false

	i := 0
	for i < len(args) {
		if strings.HasPrefix(args[i], "-") {
			if strings.HasSuffix(args[i], "-help") {
				opts.help = true
			} else if args[i] == "--repl" {
				opts.repl = true
			} else if args[i] == "--nop" {
				opts.nop = true
			} else if args[i] == "--comp" {
				opts.comp = true
			} else {
				opts.scalaArgs = append(opts.scalaArgs, args[i])
			}
			i++
		} else {
			break
		}
	}

	// next arg is logically the script to run
	//
	if i < len(args) {
		opts.scriptName = args[i]
		i++
	}

	// the rest of the arguments belong to the script
	//
	for i < len(args) {
		opts.scriptArgs = append(opts.scriptArgs, args[i])
		i++
	}

	logger.Println("help       :", opts.help)
	logger.Println("repl       :", opts.repl)
	logger.Println("nop        :", opts.nop)
	logger.Println("comp       :", opts.comp)
	logger.Println("scala.args :", opts.scalaArgs)
	logger.Println("script.name:", opts.scriptName)
	logger.Println("script.args:", opts.scriptArgs)

	// skip out on help request
	//
	if opts.help {
		usage()
		os.Exit(0)
	}

	// validate our two modes of operation
	//
	if opts.repl && opts.scriptName != "" {
		usage()
		sysLogger.Println("error:", "cannot supply both repl and script")
		os.Exit(1)
	}

	// check if the file exists
	//
	if opts.scriptName != "" {
		_, err := os.Stat(opts.scriptName)
		if err != nil {
			usage()
			sysLogger.Println("error:", "cannot locate script", opts.scriptName)
			os.Exit(1)
		}
	}

	return opts
}

// entry point
//
func main() {
	args := os.Args[1:]
	opts := parse(args)

	cmdFile, err := exec.LookPath("scala")
	if err != nil {
		sysLogger.Println("error:", "unable to locate scala")
		os.Exit(1)
	}

	wd, err := os.Getwd()
	check(err)

	var cp string
	var result *SourceFile
	var scriptFile string
	var scriptSrcFile string

	if opts.repl {
		cp = gatherClassPath(wd)
		scriptFile = wd
		scriptSrcFile = scriptFile
	} else {
		result, scriptFile, err = loadFile(wd, opts.scriptName)
		check(err)
		scriptSrcFile = filepath.Join(result.folder, result.src)

		cp = gatherClassPath(result.folder)
	}

	launchArgs := []string{cmdFile}
	launchArgs = append(launchArgs, "-classpath")
	launchArgs = append(launchArgs, cp)
	if !opts.nop {
		launchArgs = append(launchArgs, "-deprecation")
		launchArgs = append(launchArgs, "-feature")
		launchArgs = append(launchArgs, "-save")
		launchArgs = append(launchArgs, "-Xlint:_")
	}
	if opts.comp {
		launchArgs = append(launchArgs, "-nc")
		launchArgs = append(launchArgs, "-nobootcp")
	}
	launchArgs = append(launchArgs, "-Dscala.script.name="+scriptSrcFile)

	launchArgs = append(launchArgs, opts.scalaArgs...)

	if !opts.repl {
		launchArgs = append(launchArgs, scriptFile)
	}

	launchArgs = append(launchArgs, opts.scriptArgs...)

	// spawn the call, replacing ourselves
	//
	logger.Println("command:", launchArgs)
	env := os.Environ()
	err = syscall.Exec(cmdFile, launchArgs, env)
	check(err)
}
