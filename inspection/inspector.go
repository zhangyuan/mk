package inspection

import "strings"

type Exception struct {
	Message string
}

type Rule struct {
	Do func(content string) (*Exception, bool)
}

func NewRule(f func(content string) (*Exception, bool)) *Rule {
	return &Rule{
		Do: f,
	}
}

type Inspector struct {
	rules []Rule
}

func NewInspector() *Inspector {
	rules := []Rule{}

	rules = append(rules, *NewRule(func(content string) (*Exception, bool) {
		if strings.Contains(content, "secret") {
			return &Exception{
				Message: content,
			}, false
		} else {
			return nil, true
		}
	}))

	rules = append(rules, *NewRule(func(content string) (*Exception, bool) {
		if strings.Contains(content, "p@ssword") {
			return &Exception{Message: content}, false
		} else {
			return nil, true
		}
	}))

	return &Inspector{
		rules: rules,
	}
}

func (inspector *Inspector) InspectFileContent(content string) (*Exception, bool) {
	for _, rule := range inspector.rules {
		if exception, ok := rule.Do(content); !ok {
			return exception, false
		}
	}

	return nil, true
}

func (inspector *Inspector) InspectCommitMessage(commit string) (*Exception, bool) {
	for _, rule := range inspector.rules {
		if exception, ok := rule.Do(commit); !ok {
			return exception, false
		}
	}

	return nil, true
}
