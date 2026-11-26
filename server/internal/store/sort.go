package store

import "sort"

// severityRank задаёт порядок вывода находок в отчётах: сначала критичные,
// затем предупреждения, в конце «в порядке».
func severityRank(severity Severity) int {
	switch severity {
	case SeverityCritical:
		return 0
	case SeverityWarning:
		return 1
	case SeverityOK:
		return 2
	default:
		return 3
	}
}

// SortBySeverity упорядочивает находки по критичности, сохраняя исходный
// порядок внутри одной группы.
func SortBySeverity(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		return severityRank(findings[i].Severity) < severityRank(findings[j].Severity)
	})
}
