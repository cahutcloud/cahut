package skill

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"cahut.cloud/internal/store"
)

// querier reads, in a transaction or out of one.
type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// row is a skill of the Cahut or of a person, as its table keeps it.
type row struct {
	id          int64
	scope       Scope
	owner       int64 // the person of a personal skill; zero for a skill of the Cahut
	name        string
	everyWindow bool
	current     int64 // the id of its current version
	previous    int64 // the id of its previous version; zero when it has one version
}

// version is a version of a skill.
type version struct {
	id          int64
	number      int
	digest      string
	text        string // the SKILL.md as it came
	description string
	size        int64 // the bytes of the SKILL.md and of the files
	writtenAt   time.Time
	writtenBy   string // the handle of its author; empty once that person is deleted
	via         string
}

// lookup returns the skill name of the Cahut, when owner is zero, or of the
// person owner; nil when there is none.
func lookup(ctx context.Context, q querier, owner int64, name string) (*row, error) {
	where, args := "owner_id IS NULL AND name = ?", []any{name}
	if owner != 0 {
		where, args = "owner_id = ? AND name = ?", []any{owner, name}
	}
	var r row
	var ownerID, previous sql.NullInt64
	err := q.QueryRowContext(ctx, "SELECT id, scope, owner_id, name, every_window, current_version, previous_version FROM skills WHERE "+where, args...).
		Scan(&r.id, &r.scope, &ownerID, &r.name, &r.everyWindow, &r.current, &previous)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.owner, r.previous = ownerID.Int64, previous.Int64
	return &r, nil
}

// versionOf returns the version id.
func versionOf(ctx context.Context, q querier, id int64) (version, error) {
	var v version
	var at string
	err := q.QueryRowContext(ctx, `
		SELECT v.id, v.number, v.digest, v.text, v.description, v.size, v.written_at, coalesce(u.handle, ''), v.via
		FROM skill_versions v LEFT JOIN users u ON u.id = v.written_by WHERE v.id = ?`, id).
		Scan(&v.id, &v.number, &v.digest, &v.text, &v.description, &v.size, &at, &v.writtenBy, &v.via)
	if err != nil {
		return version{}, err
	}
	v.writtenAt, err = store.ParseTime(at)
	return v, err
}

// numberOf returns the number of the version id, and its size.
func numberOf(ctx context.Context, q querier, id int64) (number int, size int64, err error) {
	err = q.QueryRowContext(ctx, "SELECT number, size FROM skill_versions WHERE id = ?", id).Scan(&number, &size)
	return number, size, err
}

// filesOf returns the files of the version id, by path, with their
// contents.
func filesOf(ctx context.Context, q querier, id int64) ([]File, error) {
	rows, err := q.QueryContext(ctx, "SELECT path, type, content FROM skill_files WHERE version_id = ? ORDER BY path", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []File
	for rows.Next() {
		var f File
		if err := rows.Scan(&f.Path, &f.Type, &f.Content); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

// infosOf describes the files of the version id, by path.
func infosOf(ctx context.Context, q querier, id int64) ([]FileInfo, error) {
	rows, err := q.QueryContext(ctx, "SELECT path, length(content), type FROM skill_files WHERE version_id = ? ORDER BY path", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []FileInfo
	for rows.Next() {
		var f FileInfo
		if err := rows.Scan(&f.Path, &f.Size, &f.Type); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

// fileOf returns the file at path of the version id, or false.
func fileOf(ctx context.Context, q querier, id int64, path string) (File, bool, error) {
	f := File{Path: path}
	err := q.QueryRowContext(ctx, "SELECT type, content FROM skill_files WHERE version_id = ? AND path = ?", id, path).Scan(&f.Type, &f.Content)
	if errors.Is(err, sql.ErrNoRows) {
		return File{}, false, nil
	}
	return f, err == nil, err
}

// count returns how many skills the Cahut has, when owner is zero, or the
// person owner.
func count(ctx context.Context, tx *sql.Tx, owner int64) (int, error) {
	var n int
	var err error
	if owner == 0 {
		err = tx.QueryRowContext(ctx, "SELECT count(*) FROM skills WHERE owner_id IS NULL").Scan(&n)
	} else {
		err = tx.QueryRowContext(ctx, "SELECT count(*) FROM skills WHERE owner_id = ?", owner).Scan(&n)
	}
	return n, err
}

// insert adds a skill without versions, and returns its id.
func insert(ctx context.Context, tx *sql.Tx, scope Scope, owner int64, name string, now time.Time) (int64, error) {
	var ownerID sql.NullInt64
	if owner != 0 {
		ownerID = sql.NullInt64{Int64: owner, Valid: true}
	}
	res, err := tx.ExecContext(ctx, "INSERT INTO skills (scope, owner_id, name, created_at) VALUES (?, ?, ?, ?)",
		scope, ownerID, name, store.FormatTime(now))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// addVersion adds the version v of the skill id, with its files, makes it
// its current version, keeps the version that was current as its previous
// one, and deletes the one that was previous.
func addVersion(ctx context.Context, tx *sql.Tx, id int64, v version, author int64, files []File) (int64, error) {
	var by sql.NullInt64
	if author != 0 {
		by = sql.NullInt64{Int64: author, Valid: true}
	}
	res, err := tx.ExecContext(ctx, `
		INSERT INTO skill_versions (skill_id, number, digest, text, description, size, written_at, written_by, via)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, v.number, v.digest, v.text, v.description, v.size, store.FormatTime(v.writtenAt), by, v.via)
	if err != nil {
		return 0, err
	}
	vid, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	for _, f := range files {
		if _, err := tx.ExecContext(ctx, "INSERT INTO skill_files (version_id, path, type, content) VALUES (?, ?, ?, ?)",
			vid, f.Path, f.Type, f.Content); err != nil {
			return 0, err
		}
	}
	var gone sql.NullInt64
	if err := tx.QueryRowContext(ctx, "SELECT previous_version FROM skills WHERE id = ?", id).Scan(&gone); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE skills SET previous_version = current_version, current_version = ? WHERE id = ?", vid, id); err != nil {
		return 0, err
	}
	if gone.Valid {
		if _, err := tx.ExecContext(ctx, "DELETE FROM skill_versions WHERE id = ?", gone.Int64); err != nil {
			return 0, err
		}
	}
	return vid, nil
}

// reindex writes the row of the skill id in the index of the search, from
// its name and its current version.
func reindex(ctx context.Context, tx *sql.Tx, id int64) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM skill_search WHERE rowid = ?", id); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO skill_search (rowid, name, description, text)
		SELECT s.id, s.name, v.description, v.text FROM skills s JOIN skill_versions v ON v.id = s.current_version WHERE s.id = ?`, id)
	return err
}

// raise raises the revision of the catalog of the Cahut, which AIs see.
func raise(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "UPDATE skill_catalog SET revision = revision + 1 WHERE id = 1")
	return err
}
