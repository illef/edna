package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const noteFileExt = ".edna.txt"

func notesDir() string { return os.Getenv("EDNA_NOTES_DIR") }

func noteFileName(name string) (string, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) || strings.IndexByte(name, 0) >= 0 {
		return "", errors.New("invalid note name")
	}
	return name + noteFileExt, nil
}

func writeNoteFileAtomic(name string, data []byte) error {
	dir := notesDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".edna-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(data); err == nil {
		err = f.Close()
	} else {
		_ = f.Close()
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, name))
}

func handleNotes(w http.ResponseWriter, r *http.Request) {
	if notesDir() == "" {
		http.NotFound(w, r)
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" && r.Method == http.MethodGet {
		entries, err := os.ReadDir(notesDir())
		if os.IsNotExist(err) {
			entries = nil
		} else if err != nil {
			serveInternalError(w, err)
			return
		}
		names := []string{}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), noteFileExt) {
				names = append(names, strings.TrimSuffix(entry.Name(), noteFileExt))
			}
		}
		sort.Strings(names)
		serveJSONData(w, mustJSON(names))
		return
	}
	fileName, err := noteFileName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path := filepath.Join(notesDir(), fileName)
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, path)
	case http.MethodPut:
		data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 16<<20))
		if err == nil {
			err = writeNoteFileAtomic(fileName, data)
		}
		if err != nil {
			serveInternalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			serveInternalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func mustJSON(v any) []byte { data, _ := json.Marshal(v); return data }

func handleNotesMetadata(w http.ResponseWriter, r *http.Request) {
	if notesDir() == "" {
		http.NotFound(w, r)
		return
	}
	const name = "__metadata.edna.json"
	switch r.Method {
	case http.MethodGet:
		data, err := os.ReadFile(filepath.Join(notesDir(), name))
		if os.IsNotExist(err) {
			data = []byte("[]")
		} else if err != nil {
			serveInternalError(w, err)
			return
		}
		serveJSONData(w, data)
	case http.MethodPut:
		data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err == nil {
			err = writeNoteFileAtomic(name, data)
		}
		if err != nil {
			serveInternalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
