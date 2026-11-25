package neoread

import (
	"encoding/json"
	"testing"
)

const plainPublisherTitle = "Medusa: Accelerating Serverless LLM Inference with Materialization"

func TestPaperImportRemovesPublisherMarkupFromFilename(t *testing.T) {
	s := testServer(t)
	result := uploadTo(t, s, "papers", "&lt;scp&gt;Medusa:&lt;&#47;scp&gt; Accelerating Serverless LLM Inference with Materialization.pdf", testPDF(1))
	var doc Document
	if err := json.Unmarshal(result.Body.Bytes(), &doc); err != nil || doc.Title != plainPublisherTitle {
		t.Fatalf("imported title = %q, status = %d, error = %v", doc.Title, result.Code, err)
	}
	stored, err := s.Store.Document(doc.ID)
	if err != nil || stored.Title != plainPublisherTitle {
		t.Fatalf("stored title = %q, error = %v", stored.Title, err)
	}
}

func TestOpenStoreRepairsPersistedPublisherTitles(t *testing.T) {
	root := t.TempDir()
	store, err := OpenStore(root)
	if err != nil {
		t.Fatal(err)
	}
	const marked = "<scp>Medusa:</scp> Accelerating Serverless LLM Inference with Materialization"
	const escaped = "&lt;scp&gt;Medusa:&lt;/scp&gt; Accelerating Serverless LLM Inference with Materialization"
	cases := []struct{ id, source, library, title, want string }{
		{"lookup", "lookup", "papers", marked, plainPublisherTitle},
		{"file", "file", "papers", escaped, plainPublisherTitle},
		{"manual", "manual", "papers", marked, marked},
		{"book", "lookup", "books", marked, marked},
		{"plain", "lookup", "papers", "Theory & Practice", "Theory & Practice"},
	}
	for _, tc := range cases {
		metadata := jsonText(PaperMetadata{DOI: "10.1000/example", Sources: map[string]string{"title": tc.source}, Lookup: "done"})
		_, err := store.DB.Exec("INSERT INTO documents(id,type,title,author,size,created_at,last_opened_at,library,metadata) VALUES(?,'pdf',?,'',1,'','',?,?)", tc.id, tc.title, tc.library, metadata)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := store.DB.Close(); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		reopened, err := OpenStore(root)
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range cases {
			doc, err := reopened.Document(tc.id)
			if err != nil || doc.Title != tc.want || doc.Metadata.Sources["title"] != tc.source || doc.Metadata.DOI != "10.1000/example" || doc.Metadata.Lookup != "done" {
				t.Fatalf("%s: title = %q, metadata = %+v, error = %v", tc.id, doc.Title, doc.Metadata, err)
			}
		}
		if err := reopened.DB.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
