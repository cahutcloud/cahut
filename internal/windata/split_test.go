package windata

import (
	"slices"
	"testing"
)

func TestSplit(t *testing.T) {
	for _, tc := range []struct {
		sql    string
		texts  []string
		params [][]string
	}{
		{"SELECT 1", []string{"SELECT 1"}, [][]string{nil}},
		{"SELECT 1; SELECT 2;", []string{"SELECT 1", "SELECT 2"}, [][]string{nil, nil}},
		{"SELECT ';' AS a, \"x;y\" FROM [t;u] -- ;\n; ", []string{"SELECT ';' AS a, \"x;y\" FROM [t;u] -- ;"}, [][]string{nil}},
		{"SELECT 'it''s; here' /* ; */", []string{"SELECT 'it''s; here' /* ; */"}, [][]string{nil}},
		{"INSERT INTO t VALUES (:name, @city, $age, :name)", []string{"INSERT INTO t VALUES (:name, @city, $age, :name)"}, [][]string{{"name", "city", "age"}}},
		{"SELECT ':not' || :yes", []string{"SELECT ':not' || :yes"}, [][]string{{"yes"}}},
		{
			"CREATE TRIGGER t AFTER INSERT ON a BEGIN UPDATE b SET n = CASE WHEN n > 1 THEN 1 END; DELETE FROM c; END; SELECT 1",
			[]string{"CREATE TRIGGER t AFTER INSERT ON a BEGIN UPDATE b SET n = CASE WHEN n > 1 THEN 1 END; DELETE FROM c; END", "SELECT 1"},
			[][]string{nil, nil},
		},
		{
			"create temp trigger x after delete on a begin insert into b values (1); end",
			[]string{"create temp trigger x after delete on a begin insert into b values (1); end"},
			[][]string{nil},
		},
	} {
		stmts, err := split(tc.sql)
		if err != nil {
			t.Errorf("split(%q): %v", tc.sql, err)
			continue
		}
		var texts []string
		var params [][]string
		for _, s := range stmts {
			texts = append(texts, s.text)
			params = append(params, s.params)
		}
		if !slices.Equal(texts, tc.texts) {
			t.Errorf("split(%q) = %q, want %q", tc.sql, texts, tc.texts)
		}
		for i := range params {
			if i < len(tc.params) && !slices.Equal(params[i], tc.params[i]) {
				t.Errorf("split(%q): statement %d binds %q, want %q", tc.sql, i, params[i], tc.params[i])
			}
		}
	}
}

func TestSplitRefuses(t *testing.T) {
	for _, sql := range []string{"SELECT ?", "SELECT ?1", "SELECT 'open", "SELECT 1 /* open", "SELECT [open", `SELECT "open`} {
		if _, err := split(sql); err == nil || err.Code != CodeInvalid {
			t.Errorf("split(%q) = %v, want an invalid error", sql, err)
		}
	}
}

func TestSplitMarks(t *testing.T) {
	stmts, err := split("insert into t (a) values (1) returning id")
	if err != nil || len(stmts) != 1 || stmts[0].first != "INSERT" || !stmts[0].returning {
		t.Errorf("split = %+v, %v, want one INSERT that returns rows", stmts, err)
	}
}
