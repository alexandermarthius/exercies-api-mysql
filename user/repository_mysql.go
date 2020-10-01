package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/alexandermarthius/exercies-api-mysql/config"
	"github.com/alexandermarthius/exercies-api-mysql/models"
	"golang.org/x/crypto/bcrypt"
)

const (
	table          = "user"
	layoutDateTime = "2006-01-02 15:04:05"
)

func GetByUsername(ctx context.Context, username string) (models.User, error) {
	db, err := config.MySQL()

	if err != nil {
		log.Fatal(err)
	}

	var user models.User

	// Execute the query
	err = db.QueryRow("SELECT id, role_id, username, password, active FROM user where username = ?", username).Scan(&user.ID,
		&user.RoleID,
		&user.Username,
		&user.Password,
		&user.Active,
	)

	if err != nil {
		log.Fatal(err)
	}

	return user, nil
}

func GetOne(ctx context.Context, id interface{}) (models.User, error) {
	db, err := config.MySQL()

	if err != nil {
		log.Fatal(err)
	}

	var user models.User
	var createdAt, updatedAt string

	// Execute the query
	err = db.QueryRow(`
		SELECT 
			a.id,
			a.role_id,
			a.username,
			a.active,
			a.created_at,
			a.updated_at,
			b.name role_name,
			b.description role_description
		FROM user a 
		INNER JOIN Role b ON a.role_id = b.id
		WHERE a.id = ? AND a.active = 1 AND a.deleted_at IS NULL`, id).Scan(&user.ID,
		&user.RoleID,
		&user.Username,
		&user.Active,
		&createdAt,
		&updatedAt,
		&user.Role.Name,
		&user.Role.Description,
	)

	if err != nil {
		log.Fatal(err)
	}

	user.CreatedAt, err = time.Parse(layoutDateTime, createdAt)
	if err != nil {
		log.Fatal(err)
	}

	user.UpdatedAt, err = time.Parse(layoutDateTime, updatedAt)
	if err != nil {
		log.Fatal(err)
	}

	user.Role.ID = user.RoleID

	return user, nil
}

func GetAll(ctx context.Context) ([]models.User, error) {
	var users []models.User

	db, err := config.MySQL()

	if err != nil {
		log.Fatal(err)
	}

	quertText := fmt.Sprintf(`
		SELECT 
			a.id,
			a.role_id,
			a.username,
			a.active,
			a.created_at,
			a.updated_at,
			b.name role_name,
			b.description role_description
		FROM %v a 
		INNER JOIN Role b ON a.role_id = b.id
		WHERE active = 1 AND deleted_at IS NULL`, table)
	rowQuery, err := db.QueryContext(ctx, quertText)

	if err != nil {
		log.Fatal(err)
	}

	for rowQuery.Next() {
		var user models.User
		var createdAt, updatedAt string
		// var roleName, roleDescription string

		err = rowQuery.Scan(
			&user.ID,
			&user.RoleID,
			&user.Username,
			&user.Active,
			&createdAt,
			&updatedAt,
			&user.Role.Name,
			&user.Role.Description,
		)

		if err != nil {
			return nil, err
		}

		// change format string to datetime for crated_at and updated_at
		user.CreatedAt, err = time.Parse(layoutDateTime, createdAt)
		if err != nil {
			log.Fatal(err)
		}

		user.UpdatedAt, err = time.Parse(layoutDateTime, updatedAt)
		if err != nil {
			log.Fatal(err)
		}

		user.Role.ID = user.RoleID
		users = append(users, user)
	}

	return users, nil
}

// Insert
func Insert(ctx context.Context, user models.User) error {
	db, err := config.MySQL()

	if err != nil {
		log.Fatal("Cant connect to mysql", err)
	}

	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	queryText := fmt.Sprintf("INSERT INTO %v (role_id, username, password, active, created_at, updated_at) VALUES ('%v', '%v', '%v', %v, '%v', '%v')",
		table,
		user.RoleID,
		user.Username,
		user.Password,
		user.Active,
		time.Now().Format(layoutDateTime),
		time.Now().Format(layoutDateTime),
	)

	_, err = db.ExecContext(ctx, queryText)

	if err != nil {
		return err
	}

	return nil
}

func Update(ctx context.Context, user models.User) error {
	var queryText string

	db, err := config.MySQL()

	if err != nil {
		log.Fatal(err)
	}

	if user.Password != "" {
		// Hashing the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
		queryText = fmt.Sprintf("UPDATE %v SET role_id = %v, username = '%v', password = '%v', active = %v, updated_at = '%v' WHERE id = %v",
			table,
			user.RoleID,
			user.Username,
			user.Password,
			user.Active,
			time.Now().Format(layoutDateTime),
			user.ID,
		)
	} else {

		queryText = fmt.Sprintf("UPDATE %v SET role_id = %v, username = '%v', active = %v, updated_at = '%v' WHERE id = %v",
			table,
			user.RoleID,
			user.Username,
			user.Active,
			time.Now().Format(layoutDateTime),
			user.ID,
		)
	}

	_, err = db.ExecContext(ctx, queryText)

	if err != nil {
		return err
	}

	return nil
}

func Delete(ctx context.Context, user models.User) error {
	db, err := config.MySQL()

	if err != nil {
		log.Fatal(err)
	}

	queryText := fmt.Sprintf("UPDATE %v SET deleted_at = '%v' WHERE id = %v",
		table,
		time.Now().Format(layoutDateTime),
		user.ID,
	)

	result, err := db.ExecContext(ctx, queryText)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	check, err := result.RowsAffected()
	fmt.Println(check)
	if check == 0 {
		return errors.New("id tidak ditemukan")
	}

	return nil
}
