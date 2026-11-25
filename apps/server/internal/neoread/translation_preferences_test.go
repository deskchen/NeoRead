package neoread

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTranslationPromptSettingsPreserveAndReset(t *testing.T) {
	s, _, _ := translationFixture(t)
	value := "自然流畅，保留英文术语"
	c := AIConfig{Primary: "codex", TranslationPrompt: &value}
	for _, step := range []struct {
		prompt *string
		want   string
	}{{&value, value}, {nil, value}, {new(string), ""}} {
		c.TranslationPrompt = step.prompt
		out := request(t, s, "PUT", "/api/ai/config", strings.NewReader(jsonText(c)))
		if out.Code != 200 {
			t.Fatal(out.Code, out.Body.String())
		}
		stored := s.aiConfig().TranslationPrompt
		if stored == nil || *stored != step.want {
			t.Fatalf("preference not preserved/reset: %v", stored)
		}
	}
	long := strings.Repeat("字", 6001)
	c.TranslationPrompt = &long
	if out := request(t, s, "PUT", "/api/ai/config", strings.NewReader(jsonText(c))); out.Code != 400 {
		t.Fatal("unbounded preference", out.Code)
	}
}

func TestTranslationPromptAppliesToTextButNotChatOrSource(t *testing.T) {
	s, p, m := translationFixture(t)
	var prompts, systems []string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		systems = append(systems, req.Messages[0].Content)
		prompt := req.Messages[len(req.Messages)-1].Content
		prompts = append(prompts, prompt)
		if strings.Contains(prompt, "输入资料：") {
			writeAPIReply(w, translationLine(m.Blocks[0])+translationLine(m.Blocks[1]))
		} else {
			writeAPIReply(w, "划词译文")
		}
	}))
	defer provider.Close()
	configureTranslationTest(t, s, provider.URL)
	c := s.aiConfig()
	value := "自然中文，transaction 译为事务"
	c.TranslationPrompt = &value
	if err := s.writeAIConfig(c); err != nil {
		t.Fatal(err)
	}
	if err := s.processTranslation(t.Context(), &p); err != nil {
		t.Fatal(err)
	}
	g := s.taskGenerationService(c, taskTranslation)
	if _, err := generateSelectionTranslation(t.Context(), g, selectionTranslationInput{Text: "transaction"}); err != nil {
		t.Fatal(err)
	}
	if len(prompts) != 2 {
		t.Fatal(prompts)
	}
	for _, prompt := range prompts {
		if !strings.Contains(prompt, value) {
			t.Fatal("preference missing", prompt)
		}
	}
	if !strings.Contains(prompts[0], "只输出 JSONL") || !strings.Contains(prompts[0], `"source":"First sentence. Second sentence."`) {
		t.Fatal("preference replaced the format/source contract")
	}
	if len(systems) != 2 || systems[0] != translationSystemPrompt || systems[1] != readerSystemPrompt {
		t.Fatalf("body translation instructions leaked or did not reach fallback: %v", systems)
	}
	if s.taskGenerationService(c, taskChat).translationInstructions != "" {
		t.Fatal("translation preference leaked into chat")
	}
}

func TestTranslationSettingsPauseLegacyAutoJobs(t *testing.T) {
	s := testServer(t)
	response := uploadTo(t, s, "papers", "paper.pdf", testPDF(1))
	var doc Document
	if err := json.Unmarshal(response.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	job := Processing{
		DocumentID: doc.ID, Enabled: true, Phase: "translating", Status: "queued",
		Translating: &ProcessingStage{Status: "queued"},
	}
	write := func(job Processing) {
		t.Helper()
		if err := s.Store.saveProcessing(job); err != nil {
			t.Fatal(err)
		}
	}
	settings := func(body string) {
		t.Helper()
		w := request(t, s, "PUT", "/api/settings", strings.NewReader(body))
		if w.Code != http.StatusOK {
			t.Fatal(w.Body.String())
		}
	}
	checkView := func(status string, enabled, manual bool) {
		t.Helper()
		w := request(t, s, "GET", "/api/processing", nil)
		var jobs []Processing
		if err := json.Unmarshal(w.Body.Bytes(), &jobs); err != nil || len(jobs) != 1 {
			t.Fatalf("Processing response: %s (%v)", w.Body.String(), err)
		}
		if p := jobs[0]; p.Status != status || p.Enabled != enabled || p.ManualTranslation != manual {
			t.Fatalf("Unexpected projection: %+v", p)
		}
	}
	write(job)
	settings(`{"autoTranslateMaxPages":0}`)
	checkView("paused", false, true)
	job.ManualTranslationRequested = true
	write(job)
	checkView("queued", true, false)
	settings(`{"autoTranslateMaxPages":0,"translationEnabled":false}`)
	checkView("paused", false, false)
	if stored, err := s.Store.processing(doc.ID); err != nil || stored.Status != "queued" {
		t.Fatalf("Stored queue was changed: %+v (%v)", stored, err)
	}
}
