package main

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

	err = database.run(&io.PipeWriter{}, `
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

	expectedTable := &Table{
		Name: "foo",
		Columns: []*Column{
			{
				Name: "foo",
				Type: BoolType,
			},
			{
				Name: "bar",
				Type: Int32Type,
			},
			{
				Name: "baz",
				Type: StringType,
			},
		},
	}

	t1 := database.schema.getTable("foo")
	if diff := cmp.Diff(
		t1,
		expectedTable,
		cmpopts.IgnoreFields(Table{}, "ID"),
		cmpopts.IgnoreFields(Column{}, "ID"),
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

func Test_evaluateBooleanExpressionAgainstRow(t *testing.T) {
	type args struct {
		row  *deserializedRow
		expr *expression
	}
	tests := []struct {
		name    string
		args    args
		want    *result
		wantErr bool
	}{
		{
			args: args{
				row: &deserializedRow{
					Columns: []*deserializedColumn{
						{Value: float64(23), Column: &Column{Type: Int32Type, Name: "foo"}},
					},
				},
				expr: &expression{
					_type:    Operator,
					operator: "and",
					left: &expression{
						_type:    Operator,
						operator: "greater",
						left: &expression{
							_type:   Operand,
							goValue: float64(3),
						},
						right: &expression{
							_type:   Operand,
							goValue: float64(2),
						},
					},
					right: &expression{
						_type:    Operator,
						operator: "equal",
						left: &expression{
							_type:     Operand,
							valueType: "name",
							strValue:  "foo",
						},
						right: &expression{
							_type:   Operand,
							goValue: float64(23),
						},
					},
				},
			},
			want: &result{
				goValue: true,
			},
		},
		{
			args: args{
				row: &deserializedRow{
					Columns: []*deserializedColumn{
						{Value: float64(23), Column: &Column{Type: Int32Type, Name: "foo"}},
					},
				},
				expr: &expression{
					_type:    Operator,
					operator: "or",
					left: &expression{
						_type:    Operator,
						operator: "greater",
						left: &expression{
							_type:   Operand,
							goValue: float64(3),
						},
						right: &expression{
							_type:   Operand,
							goValue: float64(2),
						},
					},
					right: &expression{
						_type:    Operator,
						operator: "equal",
						left: &expression{
							_type:     Operand,
							valueType: "name",
							strValue:  "foo",
						},
						right: &expression{
							_type:   Operand,
							goValue: float64(1),
						},
					},
				},
			},
			want: &result{
				goValue: true,
			},
		},
		{
			args: args{
				row: &deserializedRow{
					Columns: []*deserializedColumn{
						{Value: float64(23), Column: &Column{Type: Int32Type, Name: "foo"}},
					},
				},
				expr: &expression{
					_type:    Operator,
					operator: "and",
					left: &expression{
						_type:    Operator,
						operator: "greater",
						left: &expression{
							_type:   Operand,
							goValue: float64(3),
						},
						right: &expression{
							_type:   Operand,
							goValue: float64(2),
						},
					},
					right: &expression{
						_type:    Operator,
						operator: "equal",
						left: &expression{
							_type:     Operand,
							valueType: "name",
							strValue:  "foo",
						},
						right: &expression{
							_type:   Operand,
							goValue: float64(1),
						},
					},
				},
			},
			want: &result{
				goValue: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateBooleanExpressionAgainstRow(tt.args.row, tt.args.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateBooleanExpressionAgainstRow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("evaluateBooleanExpressionAgainstRow() = %v, want %v", got, tt.want)
			}
		})
	}
}
