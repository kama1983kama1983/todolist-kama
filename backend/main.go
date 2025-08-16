package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

var db *sql.DB

func main() {
	// Initialize database
	var err error
	db, err = sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create tasks table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'pending'
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Set up router
	r := mux.NewRouter()
	
	// API routes
	r.HandleFunc("/api/tasks", getTasks).Methods("GET")
	r.HandleFunc("/api/tasks", createTask).Methods("POST")
	r.HandleFunc("/api/task/{id}/complete", completeTask).Methods("PUT")
	r.HandleFunc("/api/task/{id}/update", updateTask).Methods("PUT")
	r.HandleFunc("/api/task/{id}/delete", deleteTask).Methods("DELETE")

	// Serve static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/")))

	fmt.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Handler functions
func getTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, description, status FROM tasks ORDER BY id DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO tasks (title, description, status) VALUES (?, ?, ?)",
		task.Title, task.Description, task.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	task.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"message": "Task created successfully",
	})
}

func completeTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE tasks SET status = 'completed' WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"message": "Task completed successfully",
	})
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE tasks SET title = ?, description = ?, status = ? WHERE id = ?",
		task.Title, task.Description, task.Status, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"message": "Task updated successfully",
	})
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"message": "Task deleted successfully",
	})
}
