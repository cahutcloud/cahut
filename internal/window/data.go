package window

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/windata"
)

// The data of a Window lives in its own SQLite database, which a worker of
// its own opens. See docs/adr/0056-a-database-for-each-window.md.

// CopiesKept is how many copies of its database a Window keeps.
const CopiesKept = 3

// dataDir returns the folder of the database of the Window id, beside the
// folders of its versions.
func (s *Service) dataDir(id int64) string {
	return filepath.Join(s.dir, strconv.FormatInt(id, 10), "data")
}

// hasData reports whether the Window id has a database.
func (s *Service) hasData(id int64) bool {
	_, err := os.Stat(filepath.Join(s.dataDir(id), windata.File))
	return err == nil
}

// checkData runs the migrations of d that have not run yet, prepares its
// queries and checks what it makes readable, in a transaction of the
// database of the Window that it rolls back. It sets what each query does.
func (s *Service) checkData(ctx context.Context, id int64, d *Data) (windata.Checked, error) {
	checked, err := s.data.Check(ctx, s.dataDir(id), windata.Check{Migrations: d.Migrations, Queries: d.statements(), Readable: d.Readable})
	if err != nil {
		return windata.Checked{}, dataError("declaration.data", err)
	}
	for name, q := range d.Queries {
		q.Kind = checked.Kinds[name]
		d.Queries[name] = q
	}
	return checked, nil
}

// migrateData runs the migrations of d that have not run yet, after a copy
// of the database when it has one, which rollback_window_data goes back
// to.
func (s *Service) migrateData(ctx context.Context, c *operation.Call, id int64, d *Data, copyFirst bool) error {
	dir := s.dataDir(id)
	if copyFirst {
		if _, err := s.data.Copy(ctx, dir, CopiesKept); err != nil {
			return dataError("copy the database before its migrations", err)
		}
	}
	lim, err := limitsOf(ctx, c, windata.AILimits)
	if err != nil {
		return err
	}
	if _, err := s.data.Migrate(ctx, dir, lim, d.Migrations); err != nil {
		return dataError("declaration.data", err)
	}
	return s.recordData(ctx, c, id)
}

// limitsOf returns lim, with the size that the data of each Window of the
// Cahut may reach.
func limitsOf(ctx context.Context, c *operation.Call, lim windata.Limits) (windata.Limits, error) {
	size, err := settings.DataLimit(ctx, c.Tx)
	if err != nil {
		return windata.Limits{}, err
	}
	lim.MaxSize = size
	return lim, nil
}

// recordData records the size of the database of the Window id and of its
// copies, which the limit of the disk counts.
func (s *Service) recordData(ctx context.Context, c *operation.Call, id int64) error {
	dir := s.dataDir(id)
	schema, err := s.data.Schema(ctx, dir)
	if err != nil {
		return dataError("measure the database", err)
	}
	size := schema.Size
	copies, err := windata.Copies(dir)
	if err != nil {
		return err
	}
	for _, cp := range copies {
		size += cp.Size
	}
	_, err = c.Tx.ExecContext(ctx, "UPDATE windows SET data_size = ? WHERE id = ?", size, id)
	return err
}

// dataError tells the caller what went wrong in the database of a Window:
// an error of its SQL, or that it is busy. Other errors are internal.
func dataError(what string, err error) error {
	e, ok := errors.AsType[*windata.Error](err)
	if !ok {
		return err
	}
	switch e.Code {
	case windata.CodeInvalid, windata.CodeConstraint, windata.CodeTimeout:
		return operation.Invalidf("%s: %s", what, strings.TrimSuffix(e.Message, "."))
	case windata.CodeFull:
		return operation.Invalidf("%s: the database of this Window is full", what)
	case windata.CodeBusy, windata.CodeUnavailable:
		return operation.Invalidf("%s: the database of this Window is busy: try again in a moment", what)
	}
	return err
}

// DataInfo describes the database of a Window.
type DataInfo struct {
	Tables     []windata.Table `json:"tables" jsonschema:"its tables, with their columns and how many rows each holds"`
	Views      []string        `json:"views,omitempty"`
	Indexes    []string        `json:"indexes,omitempty"`
	Triggers   []string        `json:"triggers,omitempty"`
	Migrations int             `json:"migrations" jsonschema:"how many migrations of the declaration ran"`
	Size       int64           `json:"size" jsonschema:"its size, in bytes"`
	Copies     []DataCopy      `json:"copies" jsonschema:"the copies that rollback_window_data goes back to, the latest first"`
}

// DataCopy is a copy of the database of a Window.
type DataCopy struct {
	Name string    `json:"name"`
	At   time.Time `json:"at" jsonschema:"when it was made"`
	Size int64     `json:"size"`
}

// dataInfo describes the database of the Window id, or returns nil when it
// has none.
func (s *Service) dataInfo(ctx context.Context, id int64) (*DataInfo, error) {
	if !s.hasData(id) {
		return nil, nil
	}
	dir := s.dataDir(id)
	schema, err := s.data.Schema(ctx, dir)
	if err != nil {
		return nil, dataError("describe the database", err)
	}
	info := &DataInfo{Tables: schema.Tables, Views: schema.Views, Indexes: schema.Indexes, Triggers: schema.Triggers,
		Migrations: schema.Migrations, Size: schema.Size, Copies: []DataCopy{}}
	copies, err := windata.Copies(dir)
	if err != nil {
		return nil, err
	}
	for i := len(copies) - 1; i >= 0; i-- {
		info.Copies = append(info.Copies, DataCopy{Name: copies[i].Name, At: copyTime(copies[i].Name), Size: copies[i].Size})
	}
	return info, nil
}

// copyTime reads the time in the name of a copy, copy-<Unix milliseconds>.db.
func copyTime(name string) time.Time {
	ms, err := strconv.ParseInt(strings.TrimSuffix(strings.TrimPrefix(name, "copy-"), ".db"), 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.UnixMilli(ms).UTC()
}
