package prometheus

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type MetricEvidenceOptions struct {
	MaxSeries int
	Rfc3339   bool
}

func BuildMetricEvidence(series []MetricSeries, opts ...MetricEvidenceOptions) string {
	var sb strings.Builder

	opt := MetricEvidenceOptions{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if len(series) == 0 {
		return "metric_evidence: []\n"
	}

	maxSeries := opt.MaxSeries
	if maxSeries <= 0 || maxSeries > len(series) {
		maxSeries = len(series)
	}

	timeUnit := "unix"
	if opt.Rfc3339 {
		timeUnit = "rfc3339"
	}

	sb.WriteString("metric_evidence:\n")

	for i := 0; i < maxSeries; i++ {
		s := series[i]

		sb.WriteString(fmt.Sprintf("- series: %d\n", i+1))
		sb.WriteString(fmt.Sprintf("  labels: %s\n", compactLabels(s.Labels)))

		if len(s.Points) == 0 {
			sb.WriteString("  stats: {points=0}\n")
			sb.WriteString("  samples: []\n")
			continue
		}

		minV := s.Points[0].Value
		maxV := s.Points[0].Value
		sumV := 0.0

		for _, p := range s.Points {
			if p.Value < minV {
				minV = p.Value
			}
			if p.Value > maxV {
				maxV = p.Value
			}
			sumV += p.Value
		}

		first := s.Points[0]
		last := s.Points[len(s.Points)-1]
		avgV := sumV / float64(len(s.Points))
		trend := last.Value - first.Value

		sb.WriteString(fmt.Sprintf(
			"  stats: {points=%d, latest=%.4g, min=%.4g, max=%.4g, avg=%.4g, trend=%+.4g}\n",
			len(s.Points),
			last.Value,
			minV,
			maxV,
			avgV,
			trend,
		))

		sb.WriteString("  samples: ")
		sb.WriteString(compactSamples(s.Points, timeUnit))
		sb.WriteString("\n")
	}

	if len(series) > maxSeries {
		sb.WriteString(fmt.Sprintf("- omitted_series: %d\n", len(series)-maxSeries))
	}

	return sb.String()
}

func compactLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "{}"
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		if k == "__name__" {
			continue
		}

		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, labels[k]))
	}

	return "{" + strings.Join(parts, ",") + "}"
}

func compactSamples(points []MetricPoint, timeUnit string) string {
	if len(points) == 0 {
		return "[]"
	}

	var sb strings.Builder
	sb.WriteString("[")

	for i, p := range points {
		if i > 0 {
			sb.WriteString(",")
		}

		switch timeUnit {
		case "rfc3339":
			sb.WriteString(fmt.Sprintf(
				`["%s",%.4g]`,
				p.Timestamp.Format(time.RFC3339),
				p.Value,
			))
		default:
			sb.WriteString(fmt.Sprintf(
				"[%d,%.4g]",
				p.Timestamp.Unix(),
				p.Value,
			))
		}
	}

	sb.WriteString("]")
	return sb.String()
}
