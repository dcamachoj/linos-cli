package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

type genActionFile struct {
	name    string
	minArgs int
	syntax  string
	g       *genRun
}

func newGenActionFile(g *genRun, name string, minArgs int, syntax string) *genActionFile {
	return &genActionFile{
		g:       g,
		name:    name,
		minArgs: minArgs,
		syntax:  syntax,
	}
}

func (a *genActionFile) runRe(action *genTplAction,
	handler func(re *regexp.Regexp, replace string, content string) ([]byte, error),
) error {
	var err error

	if len(action.Args) != a.minArgs {
		return fmt.Errorf("syntax: %s %s. Actual arguments: %d", a.name, a.syntax, len(action.Args))
	}

	file := action.Args[0]
	search := action.Args[1]
	replace := action.Args[2]

	filename := a.g.normalizeFile(a.g.tmpFolder, file)
	replace, err = a.g.processData(a.name, replace)
	if err != nil {
		return err
	}

	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	re, err := regexp.Compile(search)
	if err != nil {
		return err
	}

	data, err = handler(re, replace, string(data))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, data, info.Mode())
	if err != nil {
		return err
	}

	return nil
}

func (a *genActionFile) runFiles(action *genTplAction,
	handler func(srcName, dstName string) error,
) error {
	var err error

	if len(action.Args) != a.minArgs {
		return fmt.Errorf("syntax: %s %s. Actual arguments: %d", a.name, a.syntax, len(action.Args))
	}

	src := action.Args[0]
	dst := action.Args[1]
	dst, err = a.g.processData(a.name, dst)
	if err != nil {
		return err
	}
	srcName := a.g.normalizeFile(a.g.tmpFolder, src)
	dstName := a.g.normalizeFile(a.g.tmpFolder, dst)

	err = handler(srcName, dstName)
	if err != nil {
		return err
	}

	return nil
}

func (a *genActionFile) runFile(action *genTplAction,
	handler func(filename string) error,
) error {
	var err error

	if len(action.Args) != a.minArgs {
		return fmt.Errorf("syntax: %s %s. Actual arguments: %d", a.name, a.syntax, len(action.Args))
	}

	src := action.Args[0]
	srcName := a.g.normalizeFile(a.g.tmpFolder, src)

	err = handler(srcName)
	if err != nil {
		return err
	}

	return nil
}
