package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const genConfigFilename = "generator.yaml"

var genCmd = &cobra.Command{
	Use:   "generate NAME TEMPLATE",
	Short: "Generate Code from folder or Git repository",
	Long:  strings.TrimSpace(genHelp),
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		currDir, err := os.Getwd()
		if err != nil {
			return err
		}
		genObj.Name = args[0]
		genObj.Template = args[1]
		genObj.dstFolder = currDir
		genObj.tplConfig = &genTplConfig{}
		genObj.Data = map[string]string{
			"name": args[0],
		}
		for _, s := range genObj.miscData {
			idx := strings.Index(s, "=")
			if idx < 0 {
				continue
			}
			sName := s[0:idx]
			sValue := s[idx+1:]
			genObj.Data[sName] = sValue
		}
		genObj.tmpFolder, err = ioutil.TempDir("", "*")
		if err != nil {
			return err
		}
		return genObj.run()
	},
}

type genRun struct {
	Name      string
	Template  string
	GitClone  bool
	Data      map[string]string
	IsDebug   bool
	tplConfig *genTplConfig
	dstFolder string
	tmpFolder string
	miscData  []string
}

type genTplConfig struct {
	Project bool              `json:"project,omitempty"`
	Data    map[string]string `json:"data,omitempty"`
	Actions []*genTplAction   `json:"actions,omitempty"`
	Files   []string          `json:"files,omitempty"`
}

type genTplAction struct {
	Type string   `json:"type,omitempty"`
	Args []string `json:"args,omitempty"`
}

var genObj = &genRun{}

func init() {
	rootCmd.AddCommand(genCmd)
	flags := genCmd.PersistentFlags()
	flags.BoolVarP(&genObj.GitClone, "git", "g", false, "Template is a git repository")
	flags.BoolVarP(&genObj.IsDebug, "verbose", "v", false, "Enable verbose output")
	flags.StringSliceVarP(&genObj.miscData, "data", "d", nil, "User Data overrides")
}

func (g *genRun) log(format string, args ...interface{}) {
	if !g.IsDebug {
		return
	}
	if len(args) > 0 {
		format = fmt.Sprintf(format, args...)
	}
	fmt.Println(format)
}
func (g *genRun) normalizeFile(root, filename string) string {
	if os.PathSeparator != '/' {
		filename = strings.ReplaceAll(filename, "/", fmt.Sprint(os.PathSeparator))
	}
	return path.Join(root, filename)
}

func (g *genRun) checkTemplateConfig() error {
	tplName := path.Join(g.tmpFolder, genConfigFilename)
	data, err := ioutil.ReadFile(tplName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	// Remove the configuration. Already read it.
	defer os.Remove(tplName)

	g.log("Reading template config...")
	err = yaml.Unmarshal(data, &g.tplConfig)
	if err != nil {
		return err
	}

	if g.tplConfig.Data == nil {
		g.tplConfig.Data = map[string]string{}
	}
	for n, v := range g.Data {
		g.tplConfig.Data[n] = v
	}
	for n, v := range g.tplConfig.Data {
		g.log("data %s=%s", n, v)
	}
	return nil
}

func (g *genRun) copyFile(src, dst, name string, info os.FileInfo) error {
	g.log("cpy %s", name)
	data, err := ioutil.ReadFile(g.normalizeFile(src, name))
	if err != nil {
		return fmt.Errorf("Reading %s %v", name, err)
	}
	err = ioutil.WriteFile(g.normalizeFile(dst, name), data, info.Mode())
	if err != nil {
		return fmt.Errorf("Writing %s %v", name, err)
	}
	return nil
}

func (g *genRun) copyDir(src, dst, name string) error {
	srcName := g.normalizeFile(src, name)
	dstName := g.normalizeFile(dst, name)
	err := os.MkdirAll(dstName, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Can't create %s. %v", dstName, err)
	}
	infos, err := ioutil.ReadDir(srcName)
	if err != nil {
		return fmt.Errorf("Reading dir %s %v", name, err)
	}
	g.log("cpy %s/", name)
	for _, s := range infos {
		if strings.TrimSpace(s.Name()) == "" {
			continue
		}
		sName := path.Join(name, s.Name())
		if s.IsDir() {
			err = g.copyDir(src, dst, sName)
		} else {
			err = g.copyFile(src, dst, sName, s)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *genRun) processFile(name string) error {
	g.log("prc %s", name)
	filename := g.normalizeFile(g.tmpFolder, name)
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("Processing file %s. %v", name, err)
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Reading %s %v", name, err)
	}

	str, err := g.processData(name, string(data))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, []byte(str), info.Mode())
	if err != nil {
		return fmt.Errorf("Writing %s %v", name, err)
	}

	return nil
}

func (g *genRun) processData(name string, text string) (string, error) {
	tpl, err := template.New("File").Parse(text)
	if err != nil {
		return "", fmt.Errorf("Parsing template %s. %v", name, err)
	}
	tplData := map[string]string{
		"type": name,
	}
	for n, v := range g.tplConfig.Data {
		tplData[n] = v
	}
	wr := &bytes.Buffer{}
	err = tpl.Execute(wr, tplData)
	if err != nil {
		return "", fmt.Errorf("Executing template %s. %v", name, err)
	}
	return wr.String(), nil
}

func (g *genRun) processFiles() error {
	var err error
	for _, f := range g.tplConfig.Files {
		err = g.processFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *genRun) processAction(action *genTplAction) error {
	g.log("act %s %s", action.Type, strings.Join(action.Args, " "))
	var err error
	switch action.Type {
	case "rename":
		err = g.actionRename(action)
	case "copy":
		err = g.actionCopy(action)
	case "delete":
		err = g.actionDelete(action)
	case "insert-after":
		err = g.actionInsertAfter(action)
	case "insert-before":
		err = g.actionInsertBefore(action)
	case "replace-all":
		err = g.actionReplaceAll(action)
	default:
		err = fmt.Errorf("Action %s not supported", action.Type)
	}
	return err
}

func (g *genRun) actionRename(action *genTplAction) error {
	var a = newGenActionFile(g, "rename", 2, "SRC DST_TEMPLATE")
	return a.runFiles(action, func(srcName, dstName string) error {
		_, err := os.Stat(srcName)
		if err != nil {
			return err
		}
		_, err = os.Stat(dstName)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		return os.Rename(srcName, dstName)
	})
}

func (g *genRun) actionCopy(action *genTplAction) error {
	var a = newGenActionFile(g, "copy", 2, "SRC DST_TEMPLATE")
	return a.runFiles(action, func(srcName, dstName string) error {
		info, err := os.Stat(srcName)
		if err != nil {
			return err
		}
		_, err = os.Stat(dstName)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		data, err := ioutil.ReadFile(srcName)
		if err != nil {
			return err
		}

		err = os.MkdirAll(path.Dir(dstName), os.ModePerm)
		if err != nil {
			return err
		}

		return ioutil.WriteFile(dstName, data, info.Mode())
	})
}

func (g *genRun) actionDelete(action *genTplAction) error {
	var a = newGenActionFile(g, "delete", 1, "FILENAME")
	return a.runFile(action, func(filename string) error {
		return os.RemoveAll(filename)
	})
}

func (g *genRun) actionInsertAfter(action *genTplAction) error {
	var a = newGenActionFile(g, "insert-after", 3, "FILE SEARCH_REGEXP LINE_TEMPLATE")
	return a.runRe(action, func(re *regexp.Regexp, replace string, content string) ([]byte, error) {
		lines := strings.Split(content, "\n")

		for k, s := range lines {
			if re.MatchString(s) {
				lines = g.actionInsertLine(lines, k+1, replace)
				break
			}
		}

		return []byte(strings.Join(lines, "\n")), nil
	})
}

func (g *genRun) actionInsertBefore(action *genTplAction) error {
	var a = newGenActionFile(g, "insert-before", 3, "FILE SEARCH_REGEXP LINE_TEMPLATE")
	return a.runRe(action, func(re *regexp.Regexp, replace string, content string) ([]byte, error) {
		lines := strings.Split(content, "\n")

		for k, s := range lines {
			if re.MatchString(s) {
				lines = g.actionInsertLine(lines, k, replace)
				break
			}
		}

		return []byte(strings.Join(lines, "\n")), nil
	})
}

func (g *genRun) actionInsertLine(lines []string, index int, line string) []string {
	n := len(lines)
	if index < 0 {
		return lines
	}
	if index > n {
		return lines
	}
	res := make([]string, n+1)
	o := 0
	for k, s := range lines {
		if k == index {
			o = 1
			res[k] = line
		}
		res[k+o] = s
	}
	return res
}

func (g *genRun) actionReplaceAll(action *genTplAction) error {
	var a = newGenActionFile(g, "replace-all", 3, "FILE SEARCH_REGEXP REPLACE_TEMPLATE")
	return a.runRe(action, func(re *regexp.Regexp, replace string, content string) ([]byte, error) {
		return []byte(re.ReplaceAllString(content, replace)), nil
	})
}

func (g *genRun) processActions() error {
	var err error
	for k, a := range g.tplConfig.Actions {
		err = g.processAction(a)
		if err != nil {
			return fmt.Errorf("Processing action %d %s. %v", k, a.Type, err)
		}
	}
	return nil
}

func (g *genRun) validate() error {
	var err error
	if g.Name == "" {
		return fmt.Errorf("NAME is required")
	}
	if g.Template == "" {
		return fmt.Errorf("Template is required")
	}
	if g.GitClone {
		// TODO: clone git repository to tmpFolder
		run := newExecRun("git", "clone", "--depth", "1", g.Template, g.tmpFolder).
			EnvParent().Std()
		err = run.Run()
	} else {
		// Copy Template folder to tmpFolder
		err = g.copyDir(g.Template, g.tmpFolder, "")
	}
	if err != nil {
		return err
	}

	err = g.checkTemplateConfig()
	if err != nil {
		return err
	}

	g.log("ISProject: %v", g.tplConfig.Project)
	if g.tplConfig.Project {
		g.dstFolder = path.Join(g.dstFolder, g.Name)
		g.log("dstFolder: %s", g.dstFolder)
		// Check whether dstFolder is empty
		_, err := os.Stat(g.dstFolder)
		if err == nil {
			return fmt.Errorf("Destination folder %s already exists", g.dstFolder)
		}
		// There is an error and it's not NotExists
		if !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

func (g *genRun) run() error {
	// Remove temp folder on exit
	defer os.RemoveAll(g.tmpFolder)

	g.log("Validating...")
	var err = g.validate()
	if err != nil {
		return err
	}

	g.log("Processing template...")
	err = g.processFiles()
	if err != nil {
		return err
	}

	g.log("Processing actions...")
	err = g.processActions()
	if err != nil {
		return err
	}

	err = g.copyDir(g.tmpFolder, g.dstFolder, "")
	if err != nil {
		return err
	}

	g.log("Done")
	return nil
}
