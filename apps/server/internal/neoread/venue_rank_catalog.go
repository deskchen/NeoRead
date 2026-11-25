package neoread

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed venue_rank_catalog.json
var localVenueRankJSON []byte

type localVenueRank struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation,omitempty"`
	Kind         string `json:"kind"`
}

var localVenueRanks = func() map[string]map[string]string {
	var sources map[string][]localVenueRank
	if err := json.Unmarshal(localVenueRankJSON, &sources); err != nil {
		panic(err)
	}
	indexes := make(map[string]map[string]string, len(sources))
	for source, entries := range sources {
		index := map[string]string{}
		for _, entry := range entries {
			store := func(key string) {
				if key == "" {
					return
				}
				if existing, found := index[key]; found && existing != entry.Kind {
					index[key] = ""
				} else if !found {
					index[key] = entry.Kind
				}
			}
			store(journalName(entry.Name))
			if entry.Kind == "conference" {
				store(normalizeConferenceName(entry.Name))
				abbreviation := normalizeConferenceName(entry.Abbreviation)
				position, indexed := conferenceIndex[abbreviation]
				if !indexed || position >= 0 {
					store(abbreviation)
				}
				if conference, ok := matchConference(entry.Name); ok {
					store(normalizeConferenceName(conference.FullName))
				} else if conference, ok := matchConference(entry.Abbreviation); ok {
					store(normalizeConferenceName(conference.FullName))
				}
			} else {
				store(journalName(entry.Abbreviation))
			}
		}
		indexes[source] = index
	}
	return indexes
}()

func localVenueRankBadges(venue, kind string) []journalRankBadge {
	badges := []journalRankBadge{}
	for _, badge := range []journalRankBadge{
		{Key: "core", Name: "CORE", Value: "A*", Year: "2026"},
		{Key: "thcpl", Name: "TH-CPL", Value: "A", Year: "2019"},
	} {
		index := localVenueRanks[badge.Key]
		matched := index[journalName(venue)]
		if matched == "" {
			normalized := normalizeConferenceName(venue)
			position, ambiguous := conferenceIndex[normalized]
			if !ambiguous || position >= 0 {
				if conference, ok := matchConference(venue); ok {
					normalized = normalizeConferenceName(conference.FullName)
				}
				if index[normalized] == "conference" {
					matched = "conference"
				}
			}
		}
		if matched != "" && (kind == "" || kind == "preprint" || strings.EqualFold(matched, kind)) {
			badges = append(badges, badge)
		}
	}
	return badges
}
