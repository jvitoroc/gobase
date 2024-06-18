package schema

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestCreateTableHappyPath(t *testing.T) {
	s := &Schema{}

	has := s.hasTable("test")
	if has {
		t.Error("hasTable returning true, but it should return false")
		return
	}

	expectedTable := &Table{
		Name: "test",
		Columns: []*Column{
			{
				Name: "column1",
				Type: BoolType,
			},
			{
				Name: "column2",
				Type: Int32Type,
			},
			{
				Name: "column3",
				Type: StringType,
			},
		},
	}

	table, err := s.CreateTable(
		"test",
		[]*NewColumn{
			{Name: "column1", Type: BoolType},
			{Name: "column2", Type: Int32Type},
			{Name: "column3", Type: StringType},
		},
	)
	if err != nil {
		t.Error(err)
		return
	}

	has = s.hasTable(table.Name)
	if !has {
		t.Error("hasTable returning false, but it should return true")
		return
	}

	t1 := s.GetTable(table.Name)
	if diff := cmp.Diff(
		t1,
		expectedTable,
		cmpopts.IgnoreFields(Table{}, "ID"),
		cmpopts.IgnoreFields(Column{}, "ID"),
		cmpopts.IgnoreUnexported(Table{}),
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
