package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type todo struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Title  string `json:"title"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Getenv("TODO_DB"), os.Stdout, os.Stderr))
}

func run(args []string, database string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "version" {
		fmt.Fprintln(stdout, "todo-bench seed")
		return 0
	}

	if len(args) == 2 && args[0] == "add" {
		title := strings.TrimSpace(args[1])
		if title == "" {
			fmt.Fprintln(stderr, "title must not be empty")
			return 1
		}
		if database == "" {
			fmt.Fprintln(stderr, "TODO_DB must be set")
			return 1
		}

		todos, err := loadTodos(database)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		id := nextID(todos)
		todos = append(todos, todo{ID: id, Status: "active", Title: title})
		if err := saveTodos(database, todos); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "added %d\n", id)
		return 0
	}

	if len(args) == 1 && args[0] == "list" {
		if database == "" {
			fmt.Fprintln(stderr, "TODO_DB must be set")
			return 1
		}
		todos, err := loadTodos(database)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		sort.Slice(todos, func(i, j int) bool {
			return todos[i].ID < todos[j].ID
		})
		for _, item := range todos {
			fmt.Fprintf(stdout, "%d\t%s\t%s\n", item.ID, item.Status, item.Title)
		}
		return 0
	}

	fmt.Fprintln(stderr, "usage: todo <add|list|done>")
	return 2
}

func loadTodos(path string) ([]todo, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var todos []todo
	if err := json.NewDecoder(file).Decode(&todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func nextID(todos []todo) int {
	next := 1
	for _, item := range todos {
		if item.ID >= next {
			next = item.ID + 1
		}
	}
	return next
}

func saveTodos(path string, todos []todo) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".todo-*")
	if err != nil {
		return err
	}
	temporaryPath := file.Name()
	defer os.Remove(temporaryPath)

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(todos); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
