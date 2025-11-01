package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Job struct {
	ID        int
	Title     string
	Company   string
	Salary    int
	CreatedAt time.Time
}

type JobsResponse struct {
	Items       []Job
	NextAfterID int
}

var db *sql.DB

func main() {
	var err error
	db, _ := sql.Open("sqlite", "jobs.db")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	createTable := `
    CREATE TABLE IF NOT EXISTS jobs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        company TEXT NOT NULL,
        salary INTEGER NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    )`

	if _, err := db.Exec(createTable); err != nil {
		log.Fatalf("create table: %v", err)
	}

	sampleData := `
    INSERT OR IGNORE INTO jobs (title, company, salary, created_at) VALUES 
    ('Senior Go Developer', 'Kolesa', 615000, datetime('now')),
    ('Python Developer', 'Kaspi', 550000, datetime('now', '-1 day')),
    ('Java Developer', 'Kolesa', 580000, datetime('now', '-2 days')),
    ('Frontend Developer', 'Arbuz', 450000, datetime('now', '-3 days'));
	('DevOps Engineer', 'ChocoFamily', 600000, datetime('now', '-4 days'));
	('Data Analyst', 'Kolesa', 500000, datetime('now', '-5 days'));
	`

	if _, err := db.Exec(sampleData); err != nil {
		log.Printf("warning: could not insert sample data: %v", err)
	}

	http.HandleFunc("/jobs", handleJobs)
	log.Printf("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	company := strings.TrimSpace(query.Get("company"))
	afterID := strings.TrimSpace(query.Get("after_id"))
	limit := strings.TrimSpace(query.Get("limit"))

	limitInt := 10
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			if l > 100 {
				limitInt = 100
			} else {
				limitInt = l
			}
		}
	}

	args := make([]interface{}, 0)
	conditions := make([]string, 0)

	baseQuery := `
        SELECT id, title, company, salary, created_at 
        FROM jobs`

	if company != "" {
		conditions = append(conditions, "company = ?")
		args = append(args, company)
	}

	if afterID != "" {
		afterIDInt, err := strconv.Atoi(afterID)
		if err == nil && afterIDInt > 0 {
			conditions = append(conditions, `
                (created_at < (SELECT created_at FROM jobs WHERE id = ?) 
                OR (created_at = (SELECT created_at FROM jobs WHERE id = ?) 
                AND id < ?))`)
			args = append(args, afterIDInt, afterIDInt, afterIDInt)
		}
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY created_at DESC, id DESC"
	baseQuery += " LIMIT ?"
	args = append(args, limitInt)

	start := time.Now()
	rows, err := db.Query(baseQuery, args...)
	queryTime := time.Since(start)
	w.Header().Set("X-Query-Time", queryTime.String())

	if err != nil {
		log.Printf("Query error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var jobs []Job
	var lastID int
	for rows.Next() {
		var job Job
		if err := rows.Scan(&job.ID, &job.Title, &job.Company, &job.Salary, &job.CreatedAt); err != nil {
			log.Printf("Row scan error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		jobs = append(jobs, job)
		lastID = job.ID
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := JobsResponse{
		Items:       jobs,
		NextAfterID: lastID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("JSON encode error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
