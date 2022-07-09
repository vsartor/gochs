package blogger

import (
	"io/ioutil"
	"regexp"
	"strings"
	"vsartor.com/gochs/log"
)

var (
	cachedTemplates map[string]string
)

func init() {
	cachedTemplates = make(map[string]string)
}

func loadTemplate(srcDir, template string) (string, error) {
	if content, isin := cachedTemplates[template]; isin {
		log.Dbg("Fetching cached template <b>%s<r>", template)
		return content, nil
	}
	log.Info("Loading template <b>%s<r>", template)

	byteContent, err := ioutil.ReadFile(srcDir + "/templates/" + template + ".html")
	if err != nil {
		return "", log.Err("failed to open template <b>%s<r>: %s", template, err.Error())
	}
	content := string(byteContent)

	r := regexp.MustCompile(`\${[a-zA-Z-]+}`)
	for _, match := range r.FindAll(byteContent, -1) {
		subtemplate := string(match[2 : len(match)-1])
		subcontent, err := loadTemplate(srcDir, subtemplate)
		if err != nil {
			return "", err
		}

		content = strings.ReplaceAll(content, "${"+subtemplate+"}", subcontent)
	}

	cachedTemplates[template] = content
	return content, nil
}
