package tasks

import (
	"testing"
)

func TestIntegrationBuilderLambdaProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-builder-lambda"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set builder-lambda property
	setTask := BuilderLambdaPropertyTask{
		App:      appName,
		Property: "lambdayml-path",
		Value:    "config/lambda.yml",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set builder-lambda property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset builder-lambda property
	unsetTask := BuilderLambdaPropertyTask{
		App:      appName,
		Property: "lambdayml-path",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset builder-lambda property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
