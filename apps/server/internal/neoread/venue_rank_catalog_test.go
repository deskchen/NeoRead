package neoread

import (
	"encoding/json"
	"net/http"
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
	if badges := localVenueRankBadges("ASPLOS Workshop", "conference"); len(badges) != 0 {
		t.Fatalf("The main conference rank was applied to a workshop: %+v", badges)
	}
}
