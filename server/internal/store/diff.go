package store

// Diff compares findings of two scans of the same URL.
// A finding is considered the same when its code and element match, or when
// code is empty and kind, title and element match.
func Diff(previous, current []Finding) *Changes {
	key := func(f Finding) string {
		if f.Code != "" {
			return f.Code + "\x00" + f.Element
		}
		return f.Kind + "\x00" + f.Title + "\x00" + f.Element
	}

	previousKeys := make(map[string]struct{}, len(previous))
	for _, finding := range previous {
		previousKeys[key(finding)] = struct{}{}
	}

	currentKeys := make(map[string]struct{}, len(current))
	for _, finding := range current {
		currentKeys[key(finding)] = struct{}{}
	}

	added := make([]Finding, 0)
	for _, finding := range current {
		if _, exists := previousKeys[key(finding)]; !exists {
			added = append(added, finding)
		}
	}

	fixed := make([]Finding, 0)
	for _, finding := range previous {
		if _, exists := currentKeys[key(finding)]; !exists {
			fixed = append(fixed, finding)
		}
	}

	remaining := 0
	for _, finding := range current {
		if _, exists := previousKeys[key(finding)]; exists {
			remaining++
		}
	}

	return &Changes{Added: added, Fixed: fixed, Remaining: remaining}
}
