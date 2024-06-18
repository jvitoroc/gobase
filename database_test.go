package main

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jvitoroc/gobase/schema"
)

func TestSimplestEndToEnd(t *testing.T) {
	database := database{}
	err := database.initialize(t.TempDir())
	if err != nil {
		t.Error(err)
		return
	}
	byt := make([]byte, 0)
	buf := bytes.NewBuffer(byt)

	err = database.run(buf, `
		CREATE TABLE foo DEFINITIONS (
			foo bool,
			bar int,
			baz string
		);
	`)
	if err != nil {
		t.Error(err)
		return
	}

	err = database.run(buf, `
		INSERT INTO foo VALUES (true, 123, "foobarbaz");
		INSERT INTO foo VALUES (true, 312, "aaa");
	`)
	if err != nil {
		t.Error(err)
		return
	}

	err = database.run(buf, `
		SELECT foo, bar, baz FROM foo WHERE foo != false AND bar > 100;
	`)
	if err != nil {
		t.Error(err)
		return
	}

	t.Error(buf)
}

func TestDatabaseCreateTable(t *testing.T) {
	database := database{}
	err := database.initialize(t.TempDir())
	if err != nil {
		t.Error(err)
		return
	}

	err = database.run(&bytes.Buffer{}, `
		CREATE TABLE foo DEFINITIONS (
			foo bool,
			bar int,
			baz string
		);
	`)
	if err != nil {
		t.Error(err)
		return
	}

	expectedTable := &schema.Table{
		Name: "foo",
		Columns: []*schema.Column{
			{
				Name: "foo",
				Type: schema.BoolType,
			},
			{
				Name: "bar",
				Type: schema.Int32Type,
			},
			{
				Name: "baz",
				Type: schema.StringType,
			},
		},
	}

	t1 := database.schema.GetTable("foo")
	if diff := cmp.Diff(
		t1,
		expectedTable,
		cmpopts.IgnoreFields(schema.Table{}, "ID"),
		cmpopts.IgnoreFields(schema.Column{}, "ID"),
		cmpopts.IgnoreUnexported(schema.Table{}),
	); diff != "" {
		t.Error(diff)
		return
	}

	if t1.ID == 0 {
		t.Error("didn't generate id for table")
		return
	}

	for _, c := range t1.Columns {
		if c.ID == 0 {
			t.Errorf("didn't generate id for column '%s'", c.Name)
			return
		}
	}
}

func Test_database_run(t *testing.T) {
	type fields struct {
		schema *schema.Schema
	}
	type args struct {
		batch string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantR   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &database{
				schema: tt.fields.schema,
			}
			r := &bytes.Buffer{}
			if err := d.run(r, tt.args.batch); (err != nil) != tt.wantErr {
				t.Errorf("database.run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotR := r.String(); gotR != tt.wantR {
				t.Errorf("database.run() = %v, want %v", gotR, tt.wantR)
			}
		})
	}
}
