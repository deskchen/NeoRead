package scholar

import "testing"

func TestCleanTitlePreservesTextInsidePublisherMarkup(t *testing.T) {
	for _, input := range []string{
		"<scp>Medusa:</scp>\n   Accelerating Serverless LLM Inference with Materialization",
		"&lt;jats:italic&gt;Medusa:&lt;/jats:italic&gt; Accelerating Serverless LLM Inference with Materialization",
	} {
		const expected = "Medusa: Accelerating Serverless LLM Inference with Materialization"
		if title := CleanTitle(input); title != expected || TitleSimilarity(input, expected) != 1 {
			t.Fatalf("Unexpected title: %q", title)
		}
	}
}
