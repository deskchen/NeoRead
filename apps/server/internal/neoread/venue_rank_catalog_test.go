package neoread

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"
)

func TestLocalVenueRankIndicatorsWithoutKey(t *testing.T) {
	s := testServer(t)
	w := uploadTo(t, s, "papers", "paper.pdf", testPDF(1))
	var doc Document
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	w = request(t, s, "PATCH", "/api/documents/"+doc.ID, strings.NewReader(
		`{"metadata":{"venue":"Proceedings of the 30th ACM International Conference on Architectural Support for Programming Languages and Operating Systems, Volume 1","itemType":"conference"}}`))
	if w.Code != http.StatusOK {
		t.Fatal(w.Body.String())
	}
	w = request(t, s, "GET", "/api/documents/"+doc.ID+"/journal-rank", nil)
	var result journalRankResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "found" || len(result.Badges) != 2 || result.Badges[0].Value != "A*" || result.Badges[1].Value != "A" {
		t.Fatalf("Unexpected ranks: %+v", result)
	}
	if badges := localVenueRankBadges("ASPLOS Workshop", "conference"); len(badges) != 2 || badges[0].Name != "Non-CORE" || badges[1].Name != "Non-THCPL" {
		t.Fatalf("The main conference rank was applied to a workshop: %+v", badges)
	}
}

func TestCCFStatusForArxivAndUnknownVenue(t *testing.T) {
	s := testServer(t)
	for _, tc := range []struct {
		name       string
		meta       PaperMetadata
		wantNonCCF bool
	}{
		{"arXiv identifier only", PaperMetadata{ArXiv: "2502.01234", ItemType: "preprint"}, true},
		{"arXiv venue", PaperMetadata{Venue: "arXiv preprint arXiv:2502.01234"}, true},
		{"unverified venue without API key", PaperMetadata{Venue: "Example Journal"}, false},
		{"formal venue with preprint identifier", PaperMetadata{Venue: "ASPLOS", ArXiv: "2502.01234"}, false},
		{"missing metadata", PaperMetadata{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := venueTestDocument(t, s, tc.meta)
			w := request(t, s, "GET", "/api/documents/"+doc.ID+"/journal-rank", nil)
			var result journalRankResult
			if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &result) != nil {
				t.Fatal(w.Code, w.Body.String())
			}
			if got := strings.Contains(w.Body.String(), "Non-CCF"); got != tc.wantNonCCF {
				t.Fatalf("unexpected CCF status: %+v", result)
			}
			if tc.wantNonCCF {
				for _, name := range []string{"Non-CORE", "Non-THCPL"} {
					if !slices.ContainsFunc(result.Badges, func(badge journalRankBadge) bool { return badge.Name == name }) {
						t.Fatalf("missing %s: %+v", name, result)
					}
				}
			}
		})
	}
}
