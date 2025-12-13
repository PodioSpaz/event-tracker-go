package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/repository"
	"github.com/PodioSpaz/event-tracker-go/internal/util"
)

// PersonRepository implements repository.PersonRepository for SQLite
type PersonRepository struct {
	db *DB
}

// NewPersonRepository creates a new PersonRepository
func NewPersonRepository(db *DB) repository.PersonRepository {
	return &PersonRepository{db: db}
}

// Create creates a new person
func (r *PersonRepository) Create(ctx context.Context, person *domain.Person) error {
	if err := person.Validate(); err != nil {
		return err
	}

	// Check for duplicate email
	if person.HasEmail() {
		exists, err := r.ExistsByEmail(ctx, person.Email)
		if err != nil {
			return err
		}
		if exists {
			return util.ErrPersonDuplicate
		}
	}

	query := `
		INSERT INTO people (first_name, last_name, email, phone)
		VALUES (?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		person.FirstName,
		person.LastName,
		person.Email,
		person.Phone,
	)

	if err != nil {
		return fmt.Errorf("failed to create person: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	person.ID = id
	person.CreatedAt = time.Now()
	person.UpdatedAt = time.Now()

	return nil
}

// GetByID retrieves a person by ID
func (r *PersonRepository) GetByID(ctx context.Context, id int64) (*domain.Person, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, created_at, updated_at
		FROM people
		WHERE id = ?
	`

	person := &domain.Person{}
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&person.ID,
		&person.FirstName,
		&person.LastName,
		&person.Email,
		&person.Phone,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, util.ErrPersonNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get person: %w", err)
	}

	// Parse timestamps
	person.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	person.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return person, nil
}

// GetAll retrieves all people
func (r *PersonRepository) GetAll(ctx context.Context) ([]*domain.Person, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, created_at, updated_at
		FROM people
		ORDER BY last_name, first_name
	`

	return r.queryPeople(ctx, query)
}

// Update updates an existing person
func (r *PersonRepository) Update(ctx context.Context, person *domain.Person) error {
	if err := person.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE people
		SET first_name = ?, last_name = ?, email = ?, phone = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		person.FirstName,
		person.LastName,
		person.Email,
		person.Phone,
		person.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update person: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return util.ErrPersonNotFound
	}

	return nil
}

// Delete deletes a person by ID
func (r *PersonRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM people WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		// Check for foreign key constraint violation
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return util.ErrAttendeeHasRegistrations
		}
		return fmt.Errorf("failed to delete person: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return util.ErrPersonNotFound
	}

	return nil
}

// FindByEmail retrieves a person by email
func (r *PersonRepository) FindByEmail(ctx context.Context, email string) (*domain.Person, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, created_at, updated_at
		FROM people
		WHERE email = ?
	`

	person := &domain.Person{}
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, util.NormalizeEmail(email)).Scan(
		&person.ID,
		&person.FirstName,
		&person.LastName,
		&person.Email,
		&person.Phone,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, util.ErrPersonNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find person by email: %w", err)
	}

	// Parse timestamps
	person.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	person.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return person, nil
}

// FindByName searches for people by name (first or last)
func (r *PersonRepository) FindByName(ctx context.Context, name string) ([]*domain.Person, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, created_at, updated_at
		FROM people
		WHERE first_name LIKE ? OR last_name LIKE ?
		ORDER BY last_name, first_name
	`

	searchTerm := "%" + name + "%"
	return r.queryPeople(ctx, query, searchTerm, searchTerm)
}

// Search searches for people by query (name or email)
func (r *PersonRepository) Search(ctx context.Context, query string) ([]*domain.Person, error) {
	sqlQuery := `
		SELECT id, first_name, last_name, email, phone, created_at, updated_at
		FROM people
		WHERE first_name LIKE ? OR last_name LIKE ? OR email LIKE ?
		ORDER BY last_name, first_name
	`

	searchTerm := "%" + query + "%"
	return r.queryPeople(ctx, sqlQuery, searchTerm, searchTerm, searchTerm)
}

// Count returns the total number of people
func (r *PersonRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM people`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count people: %w", err)
	}

	return count, nil
}

// ExistsByEmail checks if a person with the given email exists
func (r *PersonRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT COUNT(*) FROM people WHERE email = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, util.NormalizeEmail(email)).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

// queryPeople is a helper method to query people with optional parameters
func (r *PersonRepository) queryPeople(ctx context.Context, query string, args ...interface{}) ([]*domain.Person, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query people: %w", err)
	}
	defer rows.Close()

	var people []*domain.Person

	for rows.Next() {
		person := &domain.Person{}
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&person.ID,
			&person.FirstName,
			&person.LastName,
			&person.Email,
			&person.Phone,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan person: %w", err)
		}

		// Parse timestamps
		person.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		person.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		people = append(people, person)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating people: %w", err)
	}

	return people, nil
}
