package main

import (
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func generateFinding(generatorId string, status types.ComplianceStatus) *Finding {
	return &Finding{
		GeneratorId: generatorId,
		Status:      string(status),
	}
}

func generateFindingByTitle(title string, status types.ComplianceStatus) *Finding {
	return &Finding{
		Title:  title,
		Status: string(status),
	}
}

func TestCalculator(t *testing.T) {

	t.Run("No findings should resolve in a 100% score", func(t *testing.T) {
		calc := NewCalculator([]string{})
		assert.Equal(t, 100, int(calc.Score()))
		assert.Equal(t, 0, calc.ControlCount())
		assert.Equal(t, 0, calc.ControlPassedCount())
		assert.Equal(t, 0, calc.ControlFailedCount())
		assert.Equal(t, 0, calc.FindingCount())
	})

	t.Run("No findings but known control should resolve in a 100% score", func(t *testing.T) {
		calc := NewCalculator([]string{
			"control-1",
		})
		assert.Equal(t, 100, int(calc.Score()))
		assert.Equal(t, 1, calc.ControlCount())
		assert.Equal(t, 1, calc.ControlPassedCount())
		assert.Equal(t, 0, calc.ControlFailedCount())
		assert.Equal(t, 0, calc.FindingCount())
	})

	t.Run("7 passed findings should resolve in a 100% score", func(t *testing.T) {
		calc := NewCalculator([]string{})
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-3", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-4", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-5", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-6", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-7", types.ComplianceStatusPassed), "GeneratorId")
		assert.Equal(t, 100, int(calc.Score()))
		assert.Equal(t, 7, calc.ControlCount())
		assert.Equal(t, 7, calc.ControlPassedCount())
		assert.Equal(t, 0, calc.ControlFailedCount())
		assert.Equal(t, 7, calc.FindingCount())
	})

	t.Run("4 passed, 1 failed, 1 warning, 1 not available should resolve in a 71% score", func(t *testing.T) {
		calc := NewCalculator([]string{})
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-3", types.ComplianceStatusFailed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-4", types.ComplianceStatusWarning), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-5", types.ComplianceStatusNotAvailable), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-6", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-7", types.ComplianceStatusPassed), "GeneratorId")
		assert.Equal(t, 71, int(calc.Score()))
		assert.Equal(t, 7, calc.ControlCount())
		assert.Equal(t, 5, calc.ControlPassedCount())
		assert.Equal(t, 2, calc.ControlFailedCount())
		assert.Equal(t, 7, calc.FindingCount())
	})

	t.Run("2 controls each 2 findings, one has a warning should resolve in 50%", func(t *testing.T) {
		calc := NewCalculator([]string{})
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusWarning), "GeneratorId")
		assert.Equal(t, 50, int(calc.Score()))
		assert.Equal(t, 2, calc.ControlCount())
		assert.Equal(t, 1, calc.ControlPassedCount())
		assert.Equal(t, 1, calc.ControlFailedCount())
		assert.Equal(t, 4, calc.FindingCount())
	})

	t.Run("2 controls each 2 findings, one has a warning should resolve in 50% (different order)", func(t *testing.T) {
		calc := NewCalculator([]string{})
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusWarning), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusPassed), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusPassed), "GeneratorId")
		assert.Equal(t, 50, int(calc.Score()))
		assert.Equal(t, 2, calc.ControlCount())
		assert.Equal(t, 1, calc.ControlPassedCount())
		assert.Equal(t, 1, calc.ControlFailedCount())
		assert.Equal(t, 4, calc.FindingCount())
	})

	t.Run("2 controls each 2 findings based on title, one has a warning should resolve in 50% (different order)", func(t *testing.T) {
		calc := NewCalculator([]string{})
		calc.ProcessFinding(generateFindingByTitle("control-1", types.ComplianceStatusPassed), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-2", types.ComplianceStatusWarning), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-1", types.ComplianceStatusPassed), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-2", types.ComplianceStatusPassed), "Title")
		assert.Equal(t, 50, int(calc.Score()))
		assert.Equal(t, 2, calc.ControlCount())
		assert.Equal(t, 1, calc.ControlPassedCount())
		assert.Equal(t, 1, calc.ControlFailedCount())
		assert.Equal(t, 4, calc.FindingCount())
	})

	t.Run("Expect 4 controls, but we only have findings on 2 and only 1 has a failure expect 75%", func(t *testing.T) {
		calc := NewCalculator([]string{
			"control-1",
			"control-2",
			"control-3",
			"control-4",
		})
		calc.ProcessFinding(generateFindingByTitle("control-1", types.ComplianceStatusPassed), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-2", types.ComplianceStatusWarning), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-1", types.ComplianceStatusPassed), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-2", types.ComplianceStatusPassed), "Title")
		assert.Equal(t, 75, int(calc.Score()))
		assert.Equal(t, 4, calc.ControlCount())
		assert.Equal(t, 3, calc.ControlPassedCount())
		assert.Equal(t, 1, calc.ControlFailedCount())
		assert.Equal(t, 4, calc.FindingCount())
	})

	t.Run("Handle control wildcards", func(t *testing.T) {
		calc := NewCalculator([]string{
			"control-1",
			"control-2",
			"control-3",
			"control-4",
		})
		calc.ProcessFinding(generateFindingByTitle("control-1-abc", types.ComplianceStatusPassed), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-1-def", types.ComplianceStatusWarning), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-1-ghi", types.ComplianceStatusPassed), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-1-jklm", types.ComplianceStatusPassed), "Title")
		assert.Equal(t, 75, int(calc.Score()))
		assert.Equal(t, 4, calc.ControlCount())
		assert.Equal(t, 3, calc.ControlPassedCount())
		assert.Equal(t, 1, calc.ControlFailedCount())
		assert.Equal(t, 4, calc.FindingCount())
	})

	t.Run("Handle control wildcards multiple failures", func(t *testing.T) {
		calc := NewCalculator([]string{
			"control-1",
			"control-2",
		})
		calc.ProcessFinding(generateFindingByTitle("control-1-abc", types.ComplianceStatusWarning), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-1-def", types.ComplianceStatusWarning), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-1-ghi", types.ComplianceStatusWarning), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-1-jklm", types.ComplianceStatusWarning), "Title")
		assert.Equal(t, 50, int(calc.Score()))
		assert.Equal(t, 2, calc.ControlCount())
		assert.Equal(t, 1, calc.ControlPassedCount())
		assert.Equal(t, 1, calc.ControlFailedCount())
		assert.Equal(t, 4, calc.FindingCount())
	})

	t.Run("Handle 5 control wildcards multiple failures", func(t *testing.T) {
		calc := NewCalculator([]string{
			"control-1",
			"control-2",
			"control-3",
			"control-4",
			"control-5",
		})
		calc.ProcessFinding(generateFindingByTitle("control-1-abc", types.ComplianceStatusWarning), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-1-def", types.ComplianceStatusWarning), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-2-ghi", types.ComplianceStatusWarning), "Title")
		calc.ProcessFinding(generateFindingByTitle("control-2-jklm", types.ComplianceStatusWarning), "Title")
		assert.Equal(t, 60, int(calc.Score()))
		assert.Equal(t, 5, calc.ControlCount())
		assert.Equal(t, 3, calc.ControlPassedCount())
		assert.Equal(t, 2, calc.ControlFailedCount())
		assert.Equal(t, 4, calc.FindingCount())
	})

	t.Run("Handle 5 control with generatorId multiple failures", func(t *testing.T) {
		calc := NewCalculator([]string{
			"control-1",
			"control-2",
			"control-3",
			"control-4",
			"control-5",
		})
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusWarning), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusWarning), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusWarning), "GeneratorId")
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusWarning), "GeneratorId")
		assert.Equal(t, 60, int(calc.Score()))
		assert.Equal(t, 5, calc.ControlCount())
		assert.Equal(t, 3, calc.ControlPassedCount())
		assert.Equal(t, 2, calc.ControlFailedCount())
		assert.Equal(t, 4, calc.FindingCount())
	})
}
