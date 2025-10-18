package main

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID      int     `db:"id"`
	Name    string  `db:"name"`
	Email   string  `db:"email"`
	Balance float64 `db:"balance"`
}

var ErrInsufficientFunds = errors.New("insufficient funds")

func main() {
	// Подключаемся к БД из docker-compose (порт 5433)
	dsn := "postgres://postgres:postgres@localhost:5433/practice4?sslmode=disable"

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Пул соединений
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверим соединение с ретраями (пока контейнер поднимается)
	if err := pingRetry(db, 15, time.Second); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	fmt.Println("✅ Connected to Postgres!")

	// -------- CRUD демонстрация --------
	// 1) InsertUser (через NamedExec, как просят в задании)
	newUser := User{
		Name:    "Charlie",
		Email:   fmt.Sprintf("charlie+%d@example.com", time.Now().Unix()),
		Balance: 200.00,
	}
	if err := InsertUser(db, newUser); err != nil {
		log.Fatalf("InsertUser: %v", err)
	}
	fmt.Println("Inserted Charlie")

	// 2) GetAllUsers
	users, err := GetAllUsers(db)
	if err != nil {
		log.Fatalf("GetAllUsers: %v", err)
	}
	fmt.Println("All users:")
	for _, u := range users {
		fmt.Printf("  id=%d name=%s email=%s balance=%.2f\n", u.ID, u.Name, u.Email, u.Balance)
	}

	// 3) GetUserByID (проверим id=1 - Alice из init/users.sql)
	alice, err := GetUserByID(db, 1)
	if err != nil {
		log.Fatalf("GetUserByID(1): %v", err)
	}
	fmt.Printf("User#1: %+v\n", alice)

	// -------- Транзакция TransferBalance --------
	fmt.Println("Transfer 25.50 from id=1 to id=2 ...")
	if err := TransferBalance(db, 1, 2, 25.50); err != nil {
		log.Fatalf("TransferBalance: %v", err)
	}
	fmt.Println("OK. Balances after transfer:")
	printUsers(db)
}

func pingRetry(db *sqlx.DB, attempts int, delay time.Duration) error {
	for i := 1; i <= attempts; i++ {
		if err := db.Ping(); err != nil {
			if i == attempts {
				return err
			}
			time.Sleep(delay)
			continue
		}
		return nil
	}
	return fmt.Errorf("unreachable")
}

// 1) InsertUser — через NamedExec (задание так и просит)
func InsertUser(db *sqlx.DB, user User) error {
	const q = `
		INSERT INTO users (name, email, balance)
		VALUES (:name, :email, :balance);
	`
	_, err := db.NamedExec(q, user)
	return err
}

// 2) GetAllUsers — через Select
func GetAllUsers(db *sqlx.DB) ([]User, error) {
	const q = `SELECT id, name, email, balance FROM users ORDER BY id;`
	var out []User
	if err := db.Select(&out, q); err != nil {
		return nil, err
	}
	return out, nil
}

// 3) GetUserByID — через Get
func GetUserByID(db *sqlx.DB, id int) (User, error) {
	const q = `SELECT id, name, email, balance FROM users WHERE id=$1;`
	var u User
	if err := db.Get(&u, q, id); err != nil {
		return User{}, err
	}
	return u, nil
}

// TransferBalance — транзакция перевода баланса между пользователями
func TransferBalance(db *sqlx.DB, fromID int, toID int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if fromID == toID {
		return fmt.Errorf("cannot transfer to the same user")
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	// если не было Commit — сделаем Rollback
	defer func() { _ = tx.Rollback() }()

	// Чтобы не поймать дедлок — всегда блокируем строки в одном порядке
	ids := []int{fromID, toID}
	sort.Ints(ids)

	// Блокируем обе строки пользователей (FOR UPDATE)
	type row struct {
		ID      int     `db:"id"`
		Balance float64 `db:"balance"`
	}
	var locked []row
	query := `SELECT id, balance FROM users WHERE id = $1 OR id = $2 FOR UPDATE;`
	if err := tx.Select(&locked, query, ids[0], ids[1]); err != nil {
		return err
	}
	if len(locked) != 2 {
		return fmt.Errorf("user not found")
	}

	var fromBal float64
	for _, r := range locked {
		if r.ID == fromID {
			fromBal = r.Balance
		}
	}
	if fromBal < amount {
		return ErrInsufficientFunds
	}

	// Списываем у отправителя
	if _, err := tx.Exec(`UPDATE users SET balance = balance - $1 WHERE id = $2;`, amount, fromID); err != nil {
		return err
	}
	// Начисляем получателю
	if _, err := tx.Exec(`UPDATE users SET balance = balance + $1 WHERE id = $2;`, amount, toID); err != nil {
		return err
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func printUsers(db *sqlx.DB) {
	users, err := GetAllUsers(db)
	if err != nil {
		log.Printf("printUsers: %v", err)
		return
	}
	for _, u := range users {
		fmt.Printf("  id=%d name=%s balance=%.2f\n", u.ID, u.Name, u.Balance)
	}
}
