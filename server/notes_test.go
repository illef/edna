package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNoteFileName(t *testing.T) {
	if got, err := noteFileName("회의 기록"); err != nil || got != "회의 기록.edna.txt" {
		t.Fatalf("got %q, %v", got, err)
	}
	for _, name := range []string{"", ".", "..", "../secret", `a\b`} {
		if _, err := noteFileName(name); err == nil {
			t.Fatalf("accepted unsafe name %q", name)
		}
	}
}

func TestNotesCRUD(t *testing.T) {
	t.Setenv("EDNA_NOTES_DIR", t.TempDir())
	request := func(method, target, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, target, strings.NewReader(body))
		w := httptest.NewRecorder()
		handleNotes(w, r)
		return w
	}
	if status := request(http.MethodPut, "/api/notes?name=test", "\n∞∞∞markdown\nhello").Code; status != http.StatusNoContent {
		t.Fatalf("PUT status %d", status)
	}
	w := request(http.MethodGet, "/api/notes?name=test", "")
	body, _ := io.ReadAll(w.Result().Body)
	if string(body) != "\n∞∞∞markdown\nhello" {
		t.Fatalf("GET body %q", body)
	}
	if status := request(http.MethodDelete, "/api/notes?name=test", "").Code; status != http.StatusNoContent {
		t.Fatalf("DELETE status %d", status)
	}
}
