package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type listing struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Mtime string `json:"mtime"`
	Size  int64  `json:"size,omitempty"`
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	root, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("starting volume on %s root %s\n", port, root)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		target, err := resolve(root, r.URL.Path)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		switch r.Method {
		case "PUT":
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				w.WriteHeader(500)
				return
			}
			f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			_, copyErr := io.Copy(f, r.Body)
			closeErr := f.Close()
			if copyErr != nil || closeErr != nil {
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(201)
		case "DELETE":
			if err := os.Remove(target); err != nil {
				if os.IsNotExist(err) {
					w.WriteHeader(404)
					return
				}
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(204)
		case "GET", "HEAD":
			st, err := os.Stat(target)
			if err != nil {
				w.WriteHeader(404)
				return
			}
			if st.IsDir() {
				if r.Method == "HEAD" {
					w.WriteHeader(200)
					return
				}
				writeIndex(w, target)
				return
			}
			http.ServeFile(w, r, target)
		default:
			w.WriteHeader(405)
		}
	})
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func resolve(root, urlPath string) (string, error) {
	rel := strings.TrimPrefix(filepath.Clean("/"+urlPath), "/")
	target := filepath.Join(root, rel)
	if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes root")
	}
	return target, nil
}

func writeIndex(w http.ResponseWriter, dir string) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	out := make([]listing, 0, len(ents))
	for _, e := range ents {
		info, err := e.Info()
		if err != nil {
			continue
		}
		item := listing{
			Name:  e.Name(),
			Type:  "file",
			Mtime: info.ModTime().UTC().Format(time.RFC1123),
		}
		if e.IsDir() {
			item.Type = "directory"
		} else {
			item.Size = info.Size()
		}
		out = append(out, item)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
