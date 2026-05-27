package prometheus

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

var injectableLabels = map[string]bool{
	"instance": true,
	"job":      true,
}

func ParseGeneratorURL(generatorURL string) (string, error) {
	parsedURL, err := url.Parse(generatorURL)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %w", err)
	}

	return parsedURL.Query().Get("g0.expr"), nil
}

func buildMatchers(alertLabels map[string]string) ([]*labels.Matcher, error) {
	var matchers []*labels.Matcher
	for k, v := range alertLabels {
		if !injectableLabels[k] || v == "" {
			continue
		}
		m, err := labels.NewMatcher(labels.MatchEqual, k, v)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, m)
	}
	return matchers, nil
}

func injectMatchers(expr string, extra []*labels.Matcher) (string, error) {
	p := parser.NewParser(parser.Options{})
	ast, err := p.ParseExpr(expr)
	if err != nil {
		return "", fmt.Errorf("解析PromQL表达式失败: %w", err)
	}
	parser.Inspect(ast, func(node parser.Node, path []parser.Node) error {
		vs, ok := node.(*parser.VectorSelector)
		if !ok {
			return nil
		}
		vs.LabelMatchers, err = mergeMatchers(vs.LabelMatchers, extra)
		if err != nil {
			return fmt.Errorf("合并Matchers失败: %w", err)
		}
		return nil
	})
	return ast.String(), nil
}

func mergeMatchers(existing, extra []*labels.Matcher) ([]*labels.Matcher, error) {
	existingKeys := make(map[string]bool, len(existing))
	for _, m := range existing {
		existingKeys[m.Name] = true
	}

	result := make([]*labels.Matcher, len(existing))
	copy(result, existing)

	for _, m := range extra {
		if !existingKeys[m.Name] {
			result = append(result, m)
		}
	}
	return result, nil
}

func BuildPromQL(generatorURL string, alertLabels map[string]string) (string, error) {
	expr, err := ParseGeneratorURL(generatorURL)
	if err != nil || expr == "" {
		return "", fmt.Errorf("提取expr失败: %w", err)
	}

	// expr 里已有 instance，说明查询已经足够精准，直接返回
	if strings.Contains(expr, "instance=") {
		return expr, nil
	}

	matchers, err := buildMatchers(alertLabels)
	if err != nil {
		return "", err
	}

	return injectMatchers(expr, matchers)
}
