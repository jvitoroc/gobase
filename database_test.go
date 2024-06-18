package main

import (
	"bytes"
	"context"
	"testing"
	"time"

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
	buf := bytes.NewBuffer(make([]byte, 0))

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err = database.run(ctx, buf, `
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

	err = database.run(ctx, buf, `
		INSERT INTO foo VALUES (true, 123, "foobarbaz");
		INSERT INTO foo VALUES (true, 312, "aaa");
	`)
	if err != nil {
		t.Error(err)
		return
	}

	err = database.run(ctx, buf, `
		SELECT foo, bar, baz FROM foo WHERE foo != false AND bar > 100;
	`)
	if err != nil {
		t.Error(err)
		return
	}

	t.Error(buf.String())

	// diff := cmp.Diff(buf.String(), `{"Columns":[{"ID":1603471906,"Name":"foo","Type":"bool","Value":true},{"ID":245024461,"Name":"bar","Type":"int","Value":123},{"ID":4080717064,"Name":"baz","Type":"string","Value":"foobarbaz"}]}{"Columns":[{"ID":1603471906,"Name":"foo","Type":"bool","Value":true},{"ID":245024461,"Name":"bar","Type":"int","Value":312},{"ID":4080717064,"Name":"baz","Type":"string","Value":"aaa"}]}`)
	// if diff != "" {
	// 	t.Error(diff)
	// }
}

func TestDatabaseCreateTable(t *testing.T) {
	database := database{}
	err := database.initialize(t.TempDir())
	if err != nil {
		t.Error(err)
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err = database.run(ctx, &bytes.Buffer{}, `
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
