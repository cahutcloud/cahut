package windata

import (
	"strings"
	"sync"

	"modernc.org/sqlite"
)

// policy is what the authorizer of a connection lets a statement do while
// it is prepared. A nil policy lets the worker's own statements do
// anything; every statement that comes from a page, an AI or a person runs
// under a policy of its kind.
type policy struct {
	op       string
	readable map[string]map[string]bool // for opSelect: the readable columns of each table
	tables   map[string]bool            // for opSelect: the tables and views of the database, in lowercase
	mu       sync.Mutex
	seen     map[string]bool // the actions on rows that it saw: Adds, Changes, Deletes
	denied   string          // why it denied the last action, for the message
}

// internal reports whether a table belongs to SQLite or to Cahut. The SQL
// of pages, AIs and people never reads nor writes the tables of Cahut, and
// a page never reads the schema.
func internal(table string) bool {
	return ours(table) || strings.HasPrefix(strings.ToLower(table), "sqlite_")
}

// ours reports whether a name is kept for Cahut, such as cahut_migrations.
func ours(name string) bool {
	return strings.HasPrefix(strings.ToLower(name), "cahut_")
}

// readPragmas are the pragmas that describe the schema, which an AI may read.
var readPragmas = map[string]bool{
	"table_info": true, "table_xinfo": true, "table_list": true, "index_list": true,
	"index_info": true, "index_xinfo": true, "foreign_key_list": true,
}

// authorize decides an action, as sqlite.AuthorizerFn: see
// https://sqlite.org/c3ref/c_alter_table.html for the arguments of each.
func (p *policy) authorize(action sqlite.AuthorizerActionCode, arg1, arg2, database, _ string) sqlite.AuthorizerReturnCode {
	if p == nil {
		return sqlite.AuthorizerOK
	}
	writes := p.op == opCall || p.op == opExec
	schema := p.op == opExec || p.op == opMigrate
	db, table := database, arg1
	if action == sqlite.AuthAlterTable {
		db, table = arg1, arg2 // and database holds the column it alters
	}
	// Altering a table rewrites the schema of SQLite, the temporary one
	// included, through statements of SQLite itself.
	if db != "" && db != "main" && (db != "temp" || !schema || !internal(table)) {
		return p.deny("it reaches a database other than that of the Window")
	}
	switch action {
	case sqlite.AuthSelect, sqlite.AuthRecursive:
		return sqlite.AuthorizerOK
	case sqlite.AuthRead:
		return p.read(arg1, arg2)
	case sqlite.AuthFunction:
		name := strings.ToLower(arg2)
		if functions[name] || schema && schemaFunctions[name] {
			return sqlite.AuthorizerOK
		}
		return p.deny("the function " + arg2 + " is not allowed")
	case sqlite.AuthInsert:
		return p.row(arg1, Adds, writes || p.op == opMigrate)
	case sqlite.AuthUpdate:
		return p.row(arg1, Changes, writes)
	case sqlite.AuthDelete:
		return p.row(arg1, Deletes, writes)
	case sqlite.AuthCreateTable, sqlite.AuthCreateIndex, sqlite.AuthCreateView, sqlite.AuthCreateTrigger:
		// SQLite keeps the names that start with sqlite_ for itself, such
		// as the indexes of UNIQUE constraints; Cahut keeps those that
		// start with cahut_.
		if ours(arg1) {
			return p.deny("names that start with cahut_ are kept for Cahut")
		}
		if schema {
			return sqlite.AuthorizerOK
		}
	case sqlite.AuthAlterTable:
		if ours(arg2) {
			return p.deny("the table " + arg2 + " belongs to Cahut")
		}
		if schema {
			return sqlite.AuthorizerOK // what it removes or renames, the schema before and after tells
		}
	case sqlite.AuthReindex:
		if schema {
			return sqlite.AuthorizerOK // creating an index fills it
		}
	case sqlite.AuthDropTable, sqlite.AuthDropIndex, sqlite.AuthDropView, sqlite.AuthDropTrigger:
		if ours(arg1) {
			return p.deny(arg1 + " belongs to Cahut")
		}
		if p.op == opExec {
			return sqlite.AuthorizerOK
		}
		if p.op == opMigrate {
			return p.deny("migrations only add: remove or rename with change_window_data")
		}
	case sqlite.AuthPragma:
		if (p.op == opQuery || p.op == opExec) && readPragmas[strings.ToLower(arg1)] {
			return sqlite.AuthorizerOK
		}
		return p.deny("pragmas are not allowed")
	case sqlite.AuthTransaction, sqlite.AuthSavepoint:
		return p.deny("Cahut opens and ends the transaction of every call")
	case sqlite.AuthAttach, sqlite.AuthDetach:
		return p.deny("a Window reaches its own database only")
	}
	return p.refuse()
}

// refuse denies what the kind of the policy does not do.
func (p *policy) refuse() sqlite.AuthorizerReturnCode {
	switch p.op {
	case opSelect, opQuery:
		return p.deny("it reads only: give one SELECT")
	case opCall:
		return p.deny("a query of the declaration reads, adds, changes or deletes rows, and does nothing else")
	case opMigrate:
		return p.deny("migrations create tables, indexes, views and triggers, add columns and insert rows")
	}
	return p.deny("this statement is not allowed")
}

// row decides an action of kind on the rows of table, which the policy
// allows. Creating, altering and dropping write the schema of SQLite, which
// only those statements do: SQLite refuses every other write to it. The
// sequences of AUTOINCREMENT follow the rows of their table.
func (p *policy) row(table, kind string, allowed bool) sqlite.AuthorizerReturnCode {
	t := strings.ToLower(table)
	switch {
	case ours(t):
		return p.deny("the table " + table + " belongs to Cahut")
	case t == "sqlite_sequence":
		if allowed {
			return sqlite.AuthorizerOK
		}
	case strings.HasPrefix(t, "sqlite_"):
		if p.op == opExec || p.op == opMigrate {
			return sqlite.AuthorizerOK
		}
		return p.deny("the table " + table + " belongs to SQLite")
	case allowed:
		p.saw(kind)
		return sqlite.AuthorizerOK
	}
	return p.refuse()
}

// read decides the read of a column of a table. A reader reads the
// columns that the declaration lists, and every other column of a table it
// lists reads as NULL; a table it does not list is refused. SQLite asks
// for a column named "" when a statement reads a table without reading a
// column, as count(*) does.
func (p *policy) read(table, column string) sqlite.AuthorizerReturnCode {
	if p.op != opSelect {
		if ours(table) {
			return p.deny("the table " + table + " belongs to Cahut")
		}
		return sqlite.AuthorizerOK
	}
	cols, ok := p.readable[strings.ToLower(table)]
	if !ok {
		if !internal(table) && !p.tables[strings.ToLower(table)] {
			return sqlite.AuthorizerOK // a recursive common table expression, whose own reads are checked
		}
		return p.deny("readers do not read " + table + ": the declaration does not list it in readable")
	}
	if column == "" || cols[strings.ToLower(column)] {
		return sqlite.AuthorizerOK
	}
	return sqlite.AuthorizerIgnore
}

func (p *policy) deny(why string) sqlite.AuthorizerReturnCode {
	p.mu.Lock()
	p.denied = why
	p.mu.Unlock()
	return sqlite.AuthorizerDeny
}

func (p *policy) saw(kind string) {
	p.mu.Lock()
	if p.seen == nil {
		p.seen = map[string]bool{}
	}
	p.seen[kind] = true
	p.mu.Unlock()
}

// kind returns what a statement prepared under p does with rows.
func (p *policy) kind() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, k := range []string{Deletes, Changes, Adds} {
		if p.seen[k] {
			return k
		}
	}
	return Reads
}

func (p *policy) why() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.denied
}

// readableOf returns what the declaration makes readable, in lowercase.
func readableOf(r map[string][]string) map[string]map[string]bool {
	out := make(map[string]map[string]bool, len(r))
	for t, cols := range r {
		m := make(map[string]bool, len(cols))
		for _, c := range cols {
			m[strings.ToLower(c)] = true
		}
		out[strings.ToLower(t)] = m
	}
	return out
}

// functions are the functions that statements may call: the core
// functions of SQLite, its dates and times, its mathematics, its JSON, its
// aggregates and its windows. Every other function, such as
// load_extension, is refused. A test checks that every function SQLite
// offers is either here or in deniedFunctions.
var functions = set(
	// core
	"abs", "char", "coalesce", "concat", "concat_ws", "format", "glob", "hex", "ifnull", "iif", "if", "instr",
	"last_insert_rowid", "changes", "total_changes", "length", "like", "likelihood", "likely", "lower", "ltrim", "max", "min",
	"nullif", "octet_length", "printf", "quote", "random", "randomblob", "replace", "round", "rtrim", "sign",
	"soundex", "subtype", "substr", "substring", "trim", "typeof", "unhex", "unicode", "unlikely", "upper", "zeroblob", "unistr", "unistr_quote",
	"sqlite_version",
	// dates and times
	"date", "time", "datetime", "julianday", "unixepoch", "strftime", "timediff",
	"current_date", "current_time", "current_timestamp",
	// mathematics
	"acos", "acosh", "asin", "asinh", "atan", "atan2", "atanh", "ceil", "ceiling", "cos", "cosh", "degrees",
	"exp", "floor", "ln", "log", "log10", "log2", "mod", "pi", "pow", "power", "radians", "sin", "sinh",
	"sqrt", "tan", "tanh", "trunc",
	// JSON
	"json", "jsonb", "json_array", "jsonb_array", "json_array_insert", "jsonb_array_insert", "json_array_length", "json_error_position", "json_extract",
	"jsonb_extract", "json_insert", "jsonb_insert", "json_object", "jsonb_object", "json_patch", "jsonb_patch",
	"json_pretty", "json_remove", "jsonb_remove", "json_replace", "jsonb_replace", "json_set", "jsonb_set",
	"json_type", "json_valid", "json_quote", "json_group_array", "jsonb_group_array", "json_group_object",
	"jsonb_group_object", "json_each", "json_tree", "jsonb_each", "jsonb_tree", "->", "->>",
	// aggregates
	"avg", "count", "group_concat", "string_agg", "sum", "total", "median", "percentile", "percentile_cont", "percentile_disc",
	// windows
	"row_number", "rank", "dense_rank", "percent_rank", "cume_dist", "ntile", "lag", "lead",
	"first_value", "last_value", "nth_value",
)

// schemaFunctions are the functions of SQLite that its own statements call
// when a table is altered: nothing else can call them.
var schemaFunctions = set("sqlite_rename_column", "sqlite_rename_table", "sqlite_rename_test",
	"sqlite_rename_quotefix", "sqlite_drop_column")

// deniedFunctions are the functions that SQLite offers and statements may
// not call: they load code, read files, or tell what the build or the
// connection holds.
var deniedFunctions = set(
	"load_extension", "fts3_tokenizer", "sqlite_compileoption_get", "sqlite_compileoption_used",
	"sqlite_source_id", "sqlite_log", "sqlite_offset", "sqlite_rename_column", "sqlite_rename_table",
	"sqlite_rename_test", "sqlite_rename_quotefix", "sqlite_drop_column", "sqlite_fts3_hash",
	"fts5", "fts5_source_id", "fts5_locale", "fts5_insttoken", "fts5_get_locale",
	"match", "highlight", "snippet", "bm25", "matchinfo", "offsets", "optimize",
	"sqlite_dbpage", "sqlite_stmt", "sqlite_dbdata", "sqlite_dbptr",
	"unknown", "noop", "implies_nonnull_row", "expr_compare", "expr_implies_expr", "affinity",
	"stat_init", "stat_push", "stat_get", "sqlite_crypt",
	"geopoly_debug", "geopoly_svg", "geopoly_area", "geopoly_blob", "geopoly_json", "geopoly_bbox", "geopoly_group_bbox",
	"geopoly_contains_point", "geopoly_overlap", "geopoly_within", "geopoly_ccw", "geopoly_regular", "geopoly_xform",
	"rtreenode", "rtreedepth", "rtreecheck", "vec_version",
)

func set(names ...string) map[string]bool {
	m := make(map[string]bool, len(names))
	for _, n := range names {
		m[n] = true
	}
	return m
}
