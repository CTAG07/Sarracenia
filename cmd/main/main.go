package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/CTAG07/Sarracenia/pkg/markov"
	"os"
)

func main() {
	db, err := initDB("database.db")
	if err != nil {
		panic(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	err = markov.SetupSchema(db)
	if err != nil {
		panic(err)
	}

	gen, err := markov.NewGenerator(db, markov.NewDefaultTokenizer())
	if err != nil {
		panic(err)
	}

	model := markov.ModelInfo{Name: "test", Order: 3}

	modelInfo, err := gen.GetModelInfo(context.Background(), model.Name)
	if err == nil {
		model = modelInfo
	} else {
		err = gen.InsertModel(context.Background(), model)
		if err != nil {
			panic(err)
		}
		model, err = gen.GetModelInfo(context.Background(), model.Name)
		if err != nil {
			model = modelInfo
		}
	}

	model, err = gen.GetModelInfo(context.Background(), "test")
	if err != nil {
		panic(err)
	}

	file, err := os.Open("test.txt")
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	err = gen.Train(context.Background(), model, file)
	if err != nil {
		panic(err)
	}

	out, err := gen.Generate(context.Background(), model, markov.WithEarlyTermination(false))
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
