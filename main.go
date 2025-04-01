package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mikerybka/brass"
	"github.com/mikerybka/util"
)

const dataDir = "/home/mike/data"

func main() {
	addHost("brass.dev", brass.NewLib)
	err := http.ListenAndServe(":4000", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func addHost[RootType any](hostname string, mknew func(id string) *RootType) {
	http.HandleFunc(fmt.Sprintf("GET /%s/{owner}", hostname), list)
	http.HandleFunc(fmt.Sprintf("POST /%s/{owner}", hostname), create(mknew))

	http.HandleFunc(fmt.Sprintf("GET /%s/{owner}/{id}", hostname), getRoot[RootType])
	http.HandleFunc(fmt.Sprintf("POST /%s/{owner}/{id}", hostname), handle[RootType])
	http.HandleFunc(fmt.Sprintf("PUT /%s/{owner}/{id}", hostname), putRoot[RootType])
	http.HandleFunc(fmt.Sprintf("PATCH /%s/{owner}/{id}", hostname), patchRoot[RootType])
	http.HandleFunc(fmt.Sprintf("DELETE /%s/{owner}/{id}", hostname), deleteRoot)

	http.HandleFunc(fmt.Sprintf("/%s/{owner}/{id}/{path...}", hostname), handle[RootType])
}

func list(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(dataDir, r.URL.Path)
	entries, _ := os.ReadDir(path)
	res := []string{}
	for _, e := range entries {
		res = append(res, e.Name())
	}
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		panic(err)
	}
}

func create[T any](mknew func(id string) *T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate a unique ID
		id := util.RandomID()
		path := filepath.Join(dataDir, r.URL.Path, id)
		for {
			// check if ID already taken
			_, err := os.Stat(path)
			if errors.Is(err, os.ErrNotExist) {
				break
			}
			id = util.RandomID()
			path = filepath.Join(dataDir, r.URL.Path, id)
		}

		// Create the object
		obj := mknew(id)

		// Encode
		b, err := json.Marshal(obj)
		if err != nil {
			panic(err)
		}

		// Save the file
		err = os.WriteFile(path, b, os.ModePerm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the response
		_, err = w.Write(b)
		if err != nil {
			panic(err)
		}
	}
}

func getRoot[T any](w http.ResponseWriter, r *http.Request) {
	// Read file
	path := filepath.Join(dataDir, r.URL.Path)
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	v := new(T)
	json.NewDecoder(f).Decode(v)

	// Write response
	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		panic(err)
	}
}

func putRoot[T any](w http.ResponseWriter, r *http.Request) {
	// Read request body
	v := new(T)
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Write file
	path := filepath.Join(dataDir, r.URL.Path)
	f, err := os.Create(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func patchRoot[T any](w http.ResponseWriter, r *http.Request) {
	// Read from file
	path := filepath.Join(dataDir, r.URL.Path)
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	v := new(T)
	json.NewDecoder(f).Decode(v)
	err = f.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Patch data
	err = json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Encode new data
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write to file
	err = os.WriteFile(path, b, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write to client
	_, err = w.Write(b)
	if err != nil {
		panic(err)
	}
}

func deleteRoot(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(dataDir, r.URL.Path)
	err := os.Remove(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handle[T any](w http.ResponseWriter, r *http.Request) {
	// Read file
	path := filepath.Join(dataDir, r.URL.Path)
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	v := new(T)
	json.Unmarshal(b, v)

	// Call ServeHTTP if v is an http.Handler
	util.OptionallyServeHTTP(v, w, r)

	// Encode new data
	b, err = json.Marshal(v)
	if err != nil {
		panic(err)
	}

	// Write file
	err = os.WriteFile(path, b, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response
	_, err = w.Write(b)
	if err != nil {
		panic(err)
	}
}
