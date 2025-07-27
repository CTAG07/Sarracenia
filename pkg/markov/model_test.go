package markov

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func TestInsertAndGetModelInfo(t *testing.T) {
	_, g := setupTestDB(t)
	ctx := context.Background()

	// Test success case
	modelInfo := ModelInfo{Name: "test_model", Order: 2}
	if err := g.InsertModel(ctx, modelInfo); err != nil {
		t.Fatalf("InsertModel() failed: %v", err)
	}

	m, err := g.GetModelInfo(ctx, "test_model")
	if err != nil {
		t.Errorf("GetModelInfo: expected no error, got %v", err)
	}
	if m.Name != "test_model" || m.Order != 2 {
		t.Errorf("got unexpected model info: %+v", m)
	}

	// Test failure case (nonexistent)
	_, err = g.GetModelInfo(ctx, "nonexistent_model")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows for nonexistent model, got %v", err)
	}

	// Test failure case (duplicate name)
	err = g.InsertModel(ctx, modelInfo)
	if err == nil {
		t.Errorf("expected an error when inserting a model with a duplicate name, but got nil")
	}
}

func TestGetModelInfos(t *testing.T) {
	_, g := setupTestDB(t)
	ctx := context.Background()

	_ = g.InsertModel(ctx, ModelInfo{Name: "test_model", Order: 2})
	_ = g.InsertModel(ctx, ModelInfo{Name: "another_model", Order: 1})

	models, err := g.GetModelInfos(ctx)
	if err != nil {
		t.Fatalf("GetModelInfos failed: %v", err)
	}
	if len(models) != 2 {
		t.Errorf("expected 2 models, got %d", len(models))
	}
	if _, ok := models["test_model"]; !ok {
		t.Error("expected to find 'test_model'")
	}
	if _, ok := models["another_model"]; !ok {
		t.Error("expected to find 'another_model'")
	}
}

func TestRemoveModel(t *testing.T) {
	db, g := setupTestDB(t)
	ctx := context.Background()

	m1 := ModelInfo{Name: "to_delete", Order: 1}
	m2 := ModelInfo{Name: "to_keep", Order: 1}
	_ = g.InsertModel(ctx, m1)
	_ = g.InsertModel(ctx, m2)
	m1, _ = g.GetModelInfo(ctx, m1.Name)
	m2, _ = g.GetModelInfo(ctx, m2.Name)
	_ = g.Train(ctx, m1, strings.NewReader("delete this data."))
	_ = g.Train(ctx, m2, strings.NewReader("keep this data."))

	if err := g.RemoveModel(ctx, m1); err != nil {
		t.Fatalf("RemoveModel failed: %v", err)
	}

	// Verify model m1 is gone
	_, err := g.GetModelInfo(ctx, m1.Name)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows for deleted model, got %v", err)
	}

	// Verify chains for m1 are gone
	var count int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM markov_chains WHERE model_id = ?", m1.Id).Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 chains for deleted model, found %d", count)
	}

	// Verify model m2 and its chains still exist
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM markov_chains WHERE model_id = ?", m2.Id).Scan(&count)
	if count == 0 {
		t.Error("expected chains for kept model to exist, but found 0")
	}
}

func TestExportImportRoundTrip(t *testing.T) {
	ctx, g, modelInfo := setupTestDBWithTraining(t)

	// 1. Export the trained model to an in-memory buffer
	var buf bytes.Buffer
	if err := g.ExportModel(ctx, modelInfo, &buf); err != nil {
		t.Fatalf("ExportModel failed: %v", err)
	}

	// 2. Set up a completely new, empty database
	_, g2 := setupTestDB(t)

	// 3. Import from the buffer into the new DB
	if err := g2.ImportModel(ctx, &buf); err != nil {
		t.Fatalf("ImportModel failed: %v", err)
	}

	// 4. Verify the imported data by generating
	importedModel, err := g2.GetModelInfo(ctx, modelInfo.Name)
	if err != nil {
		t.Fatalf("could not get imported model info: %v", err)
	}

	// Generate from the imported model. The output should be predictable.
	// NOTE: It's perfectly fine for a test in model_test.go to call a Generate function
	// to verify the end-to-end success of an Import operation.
	output, err := g2.Generate(ctx, importedModel, WithMaxLength(10), WithEarlyTermination(true), WithTemperature(0))
	if err != nil {
		t.Fatalf("Generate from imported model failed: %v", err)
	}

	expected1 := "one fish two fish."
	expected2 := "red fish blue fish."
	if output != expected1 && output != expected2 {
		t.Errorf("Generate() from imported model got = %v, want one of [%q, %q]", output, expected1, expected2)
	}
}
