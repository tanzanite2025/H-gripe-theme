package suggestionfeedback

import "testing"

func TestSuggestionUploadQuotaEnforcesCountAndBytesAndReleases(t *testing.T) {
	h := NewHandler(nil, nil)
	for i := 0; i < suggestionFeedbackUploadMaxPerUserPerDay; i++ {
		if !h.reserveSuggestionUpload(7, 1<<20) {
			t.Fatalf("reserve #%d unexpectedly rejected", i+1)
		}
	}
	if h.reserveSuggestionUpload(7, 1) {
		t.Fatal("reserve beyond daily file count unexpectedly succeeded")
	}
	h.releaseSuggestionUpload(7, 1<<20)
	remaining := suggestionFeedbackUploadBytesPerUserDay - 19*(1<<20)
	if !h.reserveSuggestionUpload(7, remaining) {
		t.Fatal("reserve after release should be governed by remaining byte quota")
	}
}

func TestSuggestionUploadQuotaRejectsOversizedSingleFile(t *testing.T) {
	h := NewHandler(nil, nil)
	if h.reserveSuggestionUpload(7, suggestionFeedbackUploadBytesPerUserDay+1) {
		t.Fatal("single file larger than daily quota unexpectedly succeeded")
	}
}
