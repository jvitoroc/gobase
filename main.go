package main

import (
	"encoding/json"
	"os"
)

func main() {
	file, err := os.OpenFile("schema", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	sch := &Schema{}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(sch)
	if err != nil {
		panic(err)
	}
	defer storeSchema(sch)

	// sch.Tables[0].read()

	err = sch.Tables[0].insert([]string{"f", "0"})
	if err != nil {
		panic(err)
	}
}

func storeSchema(sch *Schema) {
	file, err := os.OpenFile("schema", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	blob, err := json.Marshal(sch)
	if err != nil {
		panic(err)
	}

	_, err = file.Write(blob)
	if err != nil {
		panic(err)
	}
}
