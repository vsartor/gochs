package blogger

import (
	"bufio"
	"fmt"
	"github.com/gomarkdown/markdown"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"vsartor.com/gochs/log"
)

var (
	cachePageSpecs  map[string]*pageSpec
	cacheGlobalVars map[string]string
)

type pageSpec struct {
	template  string
	url       string
	date      string
	title     string
	author    string
	group     string
	content   string
	preview   string
	line      string
	unlisted  bool
	variables map[string]string
}

type page struct {
	url     string
	content string
}

func buildPages(srcDir, dstDir string, prod bool) error {
	specs, err := loadPageSpecs(srcDir)
	if err != nil {
		return err
	}

	for name, spec := range specs {
		if !isDir(dstDir + "/" + spec.group) {
			err = os.Mkdir(dstDir+"/"+spec.group, os.ModePerm)
			if err != nil {
				return log.Err("failed to create folder for group <b>%s<r>: %s", spec.group, err.Error())
			}
		}

		log.Info("Loading page <b>%s<r>", name)
		page, err := loadPage(srcDir, name, *spec, prod)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(dstDir+"/"+page.url, []byte(page.content), os.ModePerm)
		if err != nil {
			return log.Err("failed to write <b>%s<r>: %s", name, err.Error())
		}
	}

	return nil
}

func loadPage(srcDir, name string, spec pageSpec, prod bool) (page, error) {
	content, err := loadTemplate(srcDir, spec.template)
	if err != nil {
		return page{}, err
	}

	content = applyVars(content, spec.variables)

	content, err = expandLists(content, srcDir)
	if err != nil {
		return page{}, err
	}

	// Apply post vars until all are filled (e.g. we want to allow "post-content" to include "post-date" macros
	r := regexp.MustCompile(`#{[a-zA-Z-]+}`)
	for r.Match([]byte(content)) {
		content = applyPostVars(content, spec)
	}

	content, err = applyGlobalVars(content, srcDir, prod)
	if err != nil {
		return page{}, err
	}

	err = checkUnfilledVars(content)
	if err != nil {
		return page{}, log.Err("unfilled variables in <b>%s<r>: %s", name, err.Error())
	}

	return page{url: spec.group + "/" + spec.url, content: content}, nil
}

func loadPageSpecs(srcDir string) (map[string]*pageSpec, error) {
	if cachePageSpecs != nil {
		return cachePageSpecs, nil
	}

	file, err := os.Open(srcDir + "/pages.spec")
	if err != nil {
		return nil, log.Err("could not open <b>pages.spec<r>: %s", err.Error())
	}
	defer file.Close()

	specs, err := parsePageSpecs(file, srcDir)
	if err != nil {
		return nil, log.Err("failed parsing page specs: %s", err.Error())
	}

	cachePageSpecs = specs
	return specs, nil
}

func applyVars(content string, vars map[string]string) string {
	for name, value := range vars {
		content = strings.ReplaceAll(content, "@{"+name+"}", value)
	}
	return content
}

func applyPostVars(content string, spec pageSpec) string {
	content = applyVars(content, spec.variables)
	content = strings.ReplaceAll(content, "#{post-title}", spec.title)
	content = strings.ReplaceAll(content, "#{post-author}", spec.author)
	content = strings.ReplaceAll(content, "#{post-date}", spec.date)
	if spec.group == "" {
		content = strings.ReplaceAll(content, "#{post-url}", spec.url)
	} else {
		content = strings.ReplaceAll(content, "#{post-url}", spec.group+"/"+spec.url)
	}

	content = strings.ReplaceAll(content, "#{post-content}", spec.content)
	content = strings.ReplaceAll(content, "#{post-preview}", spec.preview)
	content = strings.ReplaceAll(content, "#{post-line}", spec.line)

	return content
}

func applyGlobalVars(content, srcDir string, prod bool) (string, error) {
	if cacheGlobalVars == nil {
		log.Info("Loading global variables")

		file, err := os.Open(srcDir + "/globals.spec")
		if err != nil {
			return "", log.Err("could not open <b>globals.spec<r>: %s", err.Error())
		}
		defer file.Close()

		cacheGlobalVars, err = parseGlobals(file, prod)
		if err != nil {
			return "", err
		}
	}

	content = applyVars(content, cacheGlobalVars)

	return content, nil
}

func checkUnfilledVars(content string) error {
	r := regexp.MustCompile(`[$@#]{.*}`)
	match := r.Find([]byte(content))
	if match == nil {
		return nil
	}
	return fmt.Errorf("%s", match)
}

func expandLists(content string, srcDir string) (string, error) {
	bContent := []byte(content)

	rStart := regexp.MustCompile(`#{list:[A-Za-z]+:\d+}`)
	rEnd := regexp.MustCompile(`#{list:end}`)

	start := rStart.FindIndex(bContent)
	if start == nil {
		// there is no block to expand
		return content, nil
	}

	// parse start block
	toks := strings.Split(string(bContent[start[0]+2:start[1]-1]), ":")
	if len(toks) != 3 {
		panic("regexp match should assure this error does not happen")
	}
	group := toks[1]
	nReps, err := strconv.Atoi(toks[2])
	if err != nil {
		panic("regexp match should assure this error does not happen")
	}

	end := rEnd.FindIndex(bContent)
	if end == nil {
		return content, log.Err("found a start block but no end block")
	}

	// make sure the order makes sense
	if end[0] < start[1] {
		return content, log.Err("end block after the start block")
	}

	// separate bContent into: pre-start block, within blocks, post-end-block
	preContent := string(bContent[:start[0]])
	postContent := string(bContent[end[1]:])
	block := bContent[start[1]:end[0]]

	// load sorted post specs for relevant group
	allSpecs, err := loadPageSpecs(srcDir)
	if err != nil {
		panic("loadPageSpecs should never error at this point")
	}
	groupSpecs := make([]*pageSpec, 0, len(allSpecs))
	for _, spec := range allSpecs {
		if spec.group == group {
			groupSpecs = append(groupSpecs, spec)
		}
	}
	// because we want decreasing order, we switch up usual order of comparison in the "less" func
	sort.Slice(groupSpecs, func(i, j int) bool { return groupSpecs[i].date > groupSpecs[j].date })
	if nReps < len(groupSpecs) {
		groupSpecs = groupSpecs[:nReps]
	}

	// actually expand the blocks
	expandedBlocks := ""
	for _, spec := range groupSpecs {
		expandedBlocks += applyPostVars(string(block), *spec) + "\n<!-- expanded block div -->\n"
	}

	// stitch stuff together
	content = preContent + expandedBlocks + postContent

	return expandLists(content, srcDir)
}

func parsePageSpecs(file *os.File, srcDir string) (map[string]*pageSpec, error) {
	scanner := bufio.NewScanner(file)
	specs := make(map[string]*pageSpec, 0)

	currentPage := "<undef>"
	parsingVariables := false

	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " \t\n")
		if len(line) == 0 {
			continue
		}

		// handle context switching (starting new page/parsing variables)
		if line[0] == '[' {
			currentPage = line[1 : len(line)-1]
			log.Dbg("Parsing spec for <b>%s<r>", currentPage)
			specs[currentPage] = &pageSpec{}
			specs[currentPage].variables = make(map[string]string)
			parsingVariables = false
			continue
		} else if line == "variables:" {
			parsingVariables = true
			continue
		}

		// get name, value irrespective of being variable sections or not
		vals := strings.Split(line, ": ")
		if len(vals) != 2 {
			return nil, log.Err("malformatted line: %s", line)
		}
		name, value := vals[0], vals[1]

		if parsingVariables {
			specs[currentPage].variables[name] = value
		} else {
			switch name {
			case "template":
				specs[currentPage].template = value
			case "url":
				specs[currentPage].url = value
			case "title":
				specs[currentPage].title = value
			case "date":
				specs[currentPage].date = value
			case "author":
				specs[currentPage].author = value
			case "group":
				specs[currentPage].group = value
			case "unlisted":
				specs[currentPage].unlisted = true
			case "content":
				// the "value" is actually the filename we need to target to read both content and preview
				// from the "posts" and "previews" folders, respectively
				content, err := readMarkdown(srcDir + "/content/" + value)
				if err != nil {
					return nil, err
				}
				specs[currentPage].content = content

				preview, err := readMarkdown(srcDir + "/preview/" + value)
				if err != nil {
					// it's ok if a page doesn't has content but no preview
					// e.g. static pages with no "group"
					log.Warn("Skipping `preview` for <b>%s<r>", currentPage)
				} else {
					specs[currentPage].preview = preview
				}

				line, err := readMarkdown(srcDir + "/line/" + value)
				if err != nil {
					// it's ok if a page doesn't has content but no line
					// e.g. static pages with no "group"
					log.Warn("Skipping `line` for <b>%s<r>", currentPage)
				} else {
					specs[currentPage].line = line
				}
			default:
				return nil, log.Err("unexpected field: <b>%s<r>", name)
			}
		}
	}

	return specs, nil
}

func parseGlobals(file *os.File, prod bool) (map[string]string, error) {
	scanner := bufio.NewScanner(file)
	globals := make(map[string]string)

	currentVar := "<undef>"

	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " \t\n")
		if len(line) == 0 {
			continue
		}

		vals := strings.Split(line, ": ")
		if len(vals) != 2 {
			return nil, log.Err("malformatted line: %s", line)
		}
		name, value := vals[0], vals[1]

		if name == "prod" {
			if prod {
				globals[currentVar] = value
			}
		} else {
			globals[name] = value
			currentVar = name
		}
	}

	return globals, nil
}

func readMarkdown(filepath string) (string, error) {
	mdContent, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", log.Err("Failed to read md file <b>%s<r>", filepath)
	}

	htmlContent := markdown.ToHTML(mdContent, nil, nil)
	return string(htmlContent), nil
}
