package main

import (
	"reflect"
	"testing"
)

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
					columns: []*deserializedColumn{
						{value: float64(23), Column: &Column{Type: Int32Type, Name: "foo"}},
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
					columns: []*deserializedColumn{
						{value: float64(23), Column: &Column{Type: Int32Type, Name: "foo"}},
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
					columns: []*deserializedColumn{
						{value: float64(23), Column: &Column{Type: Int32Type, Name: "foo"}},
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
