package file

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/apex/log"
)

func FindFilesForTarget(dir, target string) ([]os.DirEntry, error) {
	return filesForTarget(dir, target, "file", ".yaml", false)
}

func FindScriptsForTarget(dir, target string) ([]os.DirEntry, error) {
	return filesForTarget(dir, target, "script", ".sh", true)
}

func filesForTarget(dir, target, filetype, suffix string, strict bool) ([]os.DirEntry, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	matching := map[string]os.DirEntry{}

	for _, info := range files {
		if strings.HasSuffix(info.Name(), suffix) {
			log.Debugf("considering %s '<yellow>%s</yellow>' for target: <green>%s</green>\n", filetype, info.Name(), target)
			if strings.HasSuffix(info.Name(), fmt.Sprintf("-%s%s", target, suffix)) || !strings.Contains(info.Name(), "-") {
				matching[info.Name()] = info
			} else {
				log.Debugf("not using %s '<red>%s</red>' for target: <green>%s</green>\n", filetype, info.Name(), target)
			}
		}
	}
	matchingKeys := make([]string, len(matching))
	i := 0
	for k := range matching {
		matchingKeys[i] = k
		i++
	}
	sort.Strings(matchingKeys)

	var result []os.DirEntry
	for _, k := range matchingKeys {
		if strings.HasSuffix(k, fmt.Sprintf("-%s%s", target, suffix)) {
			log.Debugf("using %s '<green>%s</green>' for target: <green>%s</green>\n", filetype, k, target)
			result = append(result, matching[k])
		} else if !strict {
			name := fmt.Sprintf("%s-%s%s", strings.TrimSuffix(k, suffix), target, suffix)
			if _, exists := matching[name]; !exists {
				log.Debugf("using %s '<green>%s</green>' for target: <green>%s</green>\n", filetype, k, target)
				result = append(result, matching[k])
			} else {
				log.Debugf("not using %s '<red>%s</red>' for target: <green>%s</green>\n", filetype, k, target)
			}
		} else {
			log.Debugf("not using %s '<red>%s</red>' for target: <green>%s</green>\n", filetype, k, target)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result, nil
}
