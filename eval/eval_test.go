package eval

import (
	"reflect"
	"testing"
)

func Test_Evaluate(t *testing.T) {
	type args struct {
		expr *Expression
		row  map[string]any
	}
	tests := []struct {
		name    string
		args    args
		want    *EvalResult
		wantErr bool
	}{
		{
			args: args{
				row: map[string]any{
					"foo": float64(23),
				},
				expr: &Expression{
					Type:     Operator,
					Operator: "and",
					Left: &Expression{
						Type:     Operator,
						Operator: "greater",
						Left: &Expression{
							Type:    Operand,
							GoValue: float64(3),
						},
						Right: &Expression{
							Type:    Operand,
							GoValue: float64(2),
						},
					},
					Right: &Expression{
						Type:     Operator,
						Operator: "equal",
						Left: &Expression{
							Type:       Operand,
							Identifier: "foo",
						},
						Right: &Expression{
							Type:    Operand,
							GoValue: float64(23),
						},
					},
				},
			},
			want: &EvalResult{
				GoValue: true,
			},
		},
		{
			args: args{
				row: map[string]any{
					"foo": float64(23),
					"bar": float64(123),
				},
				expr: &Expression{
					Type:     Operator,
					Operator: "less",
					Left: &Expression{
						Type:       Operand,
						Identifier: "bar",
					},
					Right: &Expression{
						Type:       Operand,
						Identifier: "foo",
					},
				},
			},
			want: &EvalResult{
				GoValue: false,
			},
		},
		{
			args: args{
				row: map[string]any{
					"foo": float64(123),
					"bar": float64(123),
				},
				expr: &Expression{
					Type:     Operator,
					Operator: "not_equal",
					Left: &Expression{
						Type:       Operand,
						Identifier: "bar",
					},
					Right: &Expression{
						Type:       Operand,
						Identifier: "foo",
					},
				},
			},
			want: &EvalResult{
				GoValue: false,
			},
		},
		{
			args: args{
				row: map[string]any{
					"foo": float64(12),
					"bar": float64(123),
				},
				expr: &Expression{
					Type:     Operator,
					Operator: "greater_equal",
					Left: &Expression{
						Type:       Operand,
						Identifier: "bar",
					},
					Right: &Expression{
						Type:       Operand,
						Identifier: "foo",
					},
				},
			},
			want: &EvalResult{
				GoValue: true,
			},
		},
		{
			args: args{
				row: map[string]any{
					"foo": float64(123),
					"bar": float64(123),
				},
				expr: &Expression{
					Type:     Operator,
					Operator: "less_equal",
					Left: &Expression{
						Type:       Operand,
						Identifier: "bar",
					},
					Right: &Expression{
						Type:       Operand,
						Identifier: "foo",
					},
				},
			},
			want: &EvalResult{
				GoValue: true,
			},
		},
		{
			args: args{
				row: map[string]any{
					"foo": float64(23),
				},
				expr: &Expression{
					Type:     Operator,
					Operator: "and",
					Left: &Expression{
						Type:     Operator,
						Operator: "greater",
						Left: &Expression{
							Type:    Operand,
							GoValue: float64(3),
						},
						Right: &Expression{
							Type:    Operand,
							GoValue: float64(2),
						},
					},
					Right: &Expression{
						Type:     Operator,
						Operator: "equal",
						Left: &Expression{
							Type:       Operand,
							Identifier: "foo",
						},
						Right: &Expression{
							Type:    Operand,
							GoValue: float64(1),
						},
					},
				},
			},
			want: &EvalResult{
				GoValue: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Evaluate(tt.args.expr, tt.args.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}
