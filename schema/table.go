package schema

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
)

type ColumnDataType string

const (
	StringType ColumnDataType = "string"
	BoolType   ColumnDataType = "bool"
	Int32Type  ColumnDataType = "int"
)

func checkValueType(_type ColumnDataType, value string) bool {
	switch _type {
	case BoolType:
		_, err := strconv.ParseBool(value)
		return err == nil
	case Int32Type:
		_, err := strconv.ParseInt(value, 10, 32)
		return err == nil
	case StringType:
		return true
	}

	return false
}

type Column struct {
	ID   uint32
	Name string
	Type ColumnDataType
}

func blobToGoType(_type ColumnDataType, value []byte) (any, error) {
	switch _type {
	case BoolType:
		if value[0] == 01 {
			return true, nil
		} else {
			return false, nil
		}
	case Int32Type:
		v := binary.LittleEndian.Uint32(value)
		if value[4] == 1 {
			return float64(int32(v)), nil
		} else {
			return float64(int32(v) * -1), nil
		}
	case StringType:
		return string(value), nil
	}

	return nil, errors.New("unsupported type")
}

func stringToBlob(_type ColumnDataType, value string) ([]byte, error) {
	switch _type {
	case BoolType:
		v, _ := strconv.ParseBool(value)
		if v {
			return []byte{1}, nil
		} else {
			return []byte{0}, nil
		}
	case Int32Type:
		blob := make([]byte, 5)
		v, _ := strconv.ParseInt(value, 10, 32)
		if v < 0 {
			v = v * -1
			blob[4] = 0
		} else {
			blob[4] = 1
		}
		binary.LittleEndian.PutUint32(blob, uint32(v))
		return blob, nil
	case StringType:
		return []byte(value), nil
	}

	return nil, errors.New("unsupported type")
}

type Table struct {
	ID      uint32
	Name    string
	Columns []*Column

	rootDir string
}

func (t *Table) fileName() string {
	return path.Join(t.rootDir, strconv.FormatUint(uint64(t.ID), 10))
}

func (t *Table) Insert(values []string) error {
	if len(t.Columns) != len(values) {
		return fmt.Errorf("table has %d columns, but %d values were given", len(t.Columns), len(values))
	}

	for i, c := range t.Columns {
		if ok := checkValueType(c.Type, values[i]); !ok {
			return fmt.Errorf("column '%s' data type is %s, value '%s' is invalid for this column", c.Name, c.Type, values[i])
		}
	}

	return t.write(values)
}

type DeserializedRow struct {
	Columns []*DeserializedColumn
}

func (d *DeserializedRow) GetColumn(name string) *DeserializedColumn {
	for _, c := range d.Columns {
		if c.Name == name {
			return c
		}
	}

	return nil
}

func (d *DeserializedRow) Map() map[string]any {
	m := make(map[string]any, len(d.Columns))
	for _, c := range d.Columns {
		m[c.Name] = c.Value
	}

	return m
}

func (t *Table) Read(ctx context.Context, wr io.Writer, columns []string, filter func(*DeserializedRow) (bool, error)) error {
	ch := t.createReader(ctx)

	for row := range ch {
		switch r := row.(type) {
		case []byte:
			r1, err := t.deserializeRow(r)
			if err != nil {
				wr.Write([]byte(err.Error()))
				return nil
			}

			r2, err := t.deserializeColumns(r1)
			if err != nil {
				wr.Write([]byte(err.Error()))
				return nil
			}

			b, err := json.Marshal(r2)
			if err != nil {
				wr.Write([]byte(err.Error()))
				return nil
			}

			filterResult, err := filter(r2)
			if err != nil {
				wr.Write([]byte(err.Error()))
			}

			if filterResult {
				wr.Write(b)
			}
		case error:
			wr.Write([]byte(r.Error()))
		}
	}

	return nil
}

type DeserializedColumn struct {
	*Column
	Value any
}

func (t *Table) deserializeColumns(row map[uint32][]byte) (*DeserializedRow, error) {
	r := &DeserializedRow{
		Columns: make([]*DeserializedColumn, 0, len(t.Columns)),
	}
	for _, c := range t.Columns {
		v, err := blobToGoType(c.Type, row[c.ID])
		if err != nil {
			return nil, err
		}
		r.Columns = append(r.Columns, &DeserializedColumn{
			Column: c,
			Value:  v,
		})
	}

	return r, nil
}

func (t *Table) deserializeRow(row []byte) (map[uint32][]byte, error) {
	r := bytes.NewReader(row)
	mappedRow := make(map[uint32][]byte)

	for {
		int32Bytes := make([]byte, 4)
		_, err := r.Read(int32Bytes)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		columnId := binary.LittleEndian.Uint32(int32Bytes)

		_, err = r.Read(int32Bytes)
		if err != nil {
			return nil, err
		}

		valueSize := binary.LittleEndian.Uint32(int32Bytes)

		value := make([]byte, valueSize)
		_, err = r.Read(value)
		if err != nil {
			return nil, err
		}

		mappedRow[columnId] = value
	}

	return mappedRow, nil
}

func (t *Table) createReader(ctx context.Context) chan any {
	ch := make(chan any, 1)

	go func() {
		file, err := os.OpenFile(t.fileName(), os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				ch <- ctx.Err()
				return
			default:
				rowSizeBytes := make([]byte, 4)
				n, err := file.Read(rowSizeBytes)
				if errors.Is(err, io.EOF) {
					return
				}

				if err != nil {
					ch <- err
					return
				}

				if n != 4 {
					ch <- errors.New("read invalid amount of bytes")
					return
				}

				rowSize := binary.LittleEndian.Uint32(rowSizeBytes)
				rowBytes := make([]byte, rowSize)
				n, err = file.Read(rowBytes)
				if err != nil {
					ch <- err
					return
				}

				if n != int(rowSize) {
					ch <- errors.New("read invalid amount of bytes")
					return
				}

				ch <- rowBytes
			}
		}
	}()

	return ch
}

func (t *Table) convertValuesToBlob(values []string) ([][]byte, error) {
	valuesBlob := make([][]byte, len(values))

	for i, c := range t.Columns {
		blob, err := stringToBlob(c.Type, values[i])
		if err != nil {
			return nil, err
		}

		valuesBlob[i] = blob
	}

	return valuesBlob, nil
}

func (t *Table) write(values []string) error {
	file, err := os.OpenFile(t.fileName(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var row []byte
	valuesBlob, err := t.convertValuesToBlob(values)
	if err != nil {
		return err
	}

	for i, c := range t.Columns {
		int32Bytes := make([]byte, 4)

		binary.LittleEndian.PutUint32(int32Bytes, c.ID)
		row = append(row, int32Bytes...) // write column id

		valueSize := uint32(len(valuesBlob[i]))
		binary.LittleEndian.PutUint32(int32Bytes, valueSize)
		row = append(row, int32Bytes...) // write column size

		row = append(row, valuesBlob[i]...)
	}

	rowSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(rowSizeBytes, uint32(len(row)))
	blob := append(rowSizeBytes, row...)

	_, err = file.Write(blob)
	if err != nil {
		return fmt.Errorf("an error occurred writing row to disk: %w", err)
	}

	return nil
}
