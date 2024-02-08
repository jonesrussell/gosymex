package metaresponse

import (
	"database/sql"
	"errors"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

func (r *SQLiteRepository) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS ranges (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cidr TEXT NOT NULL
	);
	`

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) Create(iprange Ipv4cidr) (*Ipv4cidr, error) {
	_, err := r.db.Exec("INSERT INTO ranges(cidr) values(?)", iprange.CIDR)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return nil, ErrDuplicate
			}
		}
		return nil, err
	}

	return &iprange, nil
}

func (r *SQLiteRepository) All() ([]Ipv4cidr, error) {
	rows, err := r.db.Query("SELECT * FROM ranges")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Ipv4cidr
	for rows.Next() {
		var iprange Ipv4cidr
		if err := rows.Scan(&iprange.CIDR); err != nil {
			return nil, err
		}
		all = append(all, iprange)
	}
	return all, nil
}

func (r *SQLiteRepository) GetByName(name string) (*Ipv4cidr, error) {
	row := r.db.QueryRow("SELECT * FROM ranges WHERE name = ?", name)

	var ranges Ipv4cidr
	if err := row.Scan(&ranges.CIDR); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	return &ranges, nil
}

func (r *SQLiteRepository) Update(id int64, updated Ipv4cidr) (*Ipv4cidr, error) {
	if id == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec("UPDATE ranges SET cidr = ? WHERE id = ?", updated.CIDR, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrUpdateFailed
	}

	return &updated, nil
}

func (r *SQLiteRepository) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM ranges WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrDeleteFailed
	}

	return err
}
