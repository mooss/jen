// Package prompts centralizes prompt definition.
package prompts

import "fmt"

var Map = map[string]string{
	"review-diff": `$(perins-jaded-review)

# Diffs

$(git diff HEAD^ HEAD)`,
	"review-staged": `$(perins-jaded-review)

# Diffs

$(git diff --staged)`,
	"commit-message": `$(per-jaded-dev)

$(ins-commit-msg)

# Diffs

$(git diff --staged)`,
	"project-graph": `$(project-graph)

${POSARGS[@]}`,
	"test": `Count the files:
$(ls)`,
}

func Raw(name string) (string, error) {
	res, present := Map[name]
	if !present {
		return "", fmt.Errorf("unknown prompt %s", name)
	}

	return res, nil
}
