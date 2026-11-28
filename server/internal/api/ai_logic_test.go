package api

import (
	"testing"

	"accessible-content/server/internal/ai"
	"accessible-content/server/internal/texts"
)

func TestImageMismatchWithLowConfidenceIsWarning(t *testing.T) {
	findings := imageFindings("mismatch", "maybe", "", "<img>", 0.2, texts.RU, true)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %+v", findings)
	}
	if findings[0].Severity != "warning" {
		t.Fatalf("low confidence must not be critical, got %q", findings[0].Severity)
	}
	if findings[0].Confidence == nil || *findings[0].Confidence != 0.2 {
		t.Fatalf("confidence must be preserved, got %+v", findings[0].Confidence)
	}
}

func TestImageMatchWithLowConfidenceIsSkipped(t *testing.T) {
	findings := imageFindings("match", "", "", "<img>", 0.2, texts.RU, true)
	if len(findings) != 0 {
		t.Fatalf("low confidence match must not add an ok finding, got %+v", findings)
	}
}

func TestImageMismatchWithHighConfidenceIsCritical(t *testing.T) {
	findings := imageFindings("mismatch", "photo shows a cat", "cat", "<img>", 0.9, texts.RU, true)
	if len(findings) != 1 || findings[0].Severity != "critical" {
		t.Fatalf("expected critical finding, got %+v", findings)
	}
}

func TestShouldReportLink(t *testing.T) {
	cases := []struct {
		name string
		res  ai.LinkResult
		want bool
	}{
		{
			name: "matches and meaningful are not a finding",
			res:  ai.LinkResult{Matches: true, Meaningful: true, SuggestedText: "better", Confidence: 0.9},
			want: false,
		},
		{
			name: "confident mismatch is a finding",
			res:  ai.LinkResult{Matches: false, Meaningful: true, Confidence: 0.9},
			want: true,
		},
		{
			name: "low confidence is skipped",
			res:  ai.LinkResult{Matches: false, Meaningful: true, Confidence: 0.2},
			want: false,
		},
		{
			name: "unclear text with high confidence is a finding",
			res:  ai.LinkResult{Matches: true, Meaningful: false, Confidence: 0.8},
			want: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := shouldReportLink(test.res); got != test.want {
				t.Fatalf("shouldReportLink() = %v, want %v", got, test.want)
			}
		})
	}
}
