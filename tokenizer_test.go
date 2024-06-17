package main

import "testing"

func Test_tokenizer_getLineColumn(t *testing.T) {
	type fields struct {
		query  string
		cursor int
		line   int
		column int
	}
	type args struct {
		skip int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantLine   int
		wantColumn int
	}{
		{
			name: "empty query, cursor overflow, skip 10",
			fields: fields{
				query:  "",
				cursor: 100,
			},
			args: args{
				skip: 10,
			},
			wantLine:   1,
			wantColumn: 1,
		},
		{
			name: "cursor overflow, skip 10",
			fields: fields{
				query:  "aaaa",
				cursor: 100,
			},
			args: args{
				skip: 10,
			},
			wantLine:   1,
			wantColumn: 5,
		},
		{
			name: "multilined, cursor overflow, skip 10",
			fields: fields{
				query:  "aaaa\nbbb\ncc",
				cursor: 100,
			},
			args: args{
				skip: 10,
			},
			wantLine:   3,
			wantColumn: 3,
		},
		{
			name: "skip 0",
			fields: fields{
				query:  "aaaa\nbbb\ncc",
				cursor: 1,
			},
			args: args{
				skip: 0,
			},
			wantLine:   1,
			wantColumn: 2,
		},
		{
			name: "first column, second line",
			fields: fields{
				query:  "aaaa\nbbb\ncc",
				cursor: 5,
			},
			args: args{
				skip: 0,
			},
			wantLine:   2,
			wantColumn: 1,
		},
		{
			name: "first column, third line, skip 3",
			fields: fields{
				query:  "aaaa\nbbb\ncc",
				cursor: 5,
			},
			args: args{
				skip: 3,
			},
			wantLine:   2,
			wantColumn: 4,
		},
		{
			name: "first column, second line, skip 4",
			fields: fields{
				query:  "aaaa\nbbb\ncc",
				cursor: 5,
			},
			args: args{
				skip: 4,
			},
			wantLine:   3,
			wantColumn: 1,
		},
		{
			name: "cursor 0, skip 0",
			fields: fields{
				query:  "aaaa\nbbb\ncc",
				cursor: 0,
			},
			args: args{
				skip: 0,
			},
			wantLine:   1,
			wantColumn: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &tokenizer{
				query:  tt.fields.query,
				cursor: tt.fields.cursor,
				line:   tt.fields.line,
				column: tt.fields.column,
			}
			got, got1 := tr.getLineColumn(tt.args.skip)
			if got != tt.wantLine {
				t.Errorf("tokenizer.getLineColumn() got line = %v, want line %v", got, tt.wantLine)
			}
			if got1 != tt.wantColumn {
				t.Errorf("tokenizer.getLineColumn() got column = %v, want column %v", got1, tt.wantColumn)
			}
		})
	}
}
