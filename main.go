package main

// func storeSchema(sch *schema.Schema) {
// 	file, err := os.OpenFile("schema", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()

// 	blob, err := json.Marshal(sch)
// 	if err != nil {
// 		panic(err)
// 	}

// 	_, err = file.Write(blob)
// 	if err != nil {
// 		panic(err)
// 	}
// }
