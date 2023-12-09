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

func TestCalculator(t *testing.T) {

	t.Run("No findings should resolve in a 100% score", func(t *testing.T) {
		calc := NewCalculator()
		assert.Equal(t, 100, int(calc.Score()))
		assert.Equal(t, 0, calc.ControlCount())
		assert.Equal(t, 0, calc.FindingCount())
	})

	t.Run("7 passed findings should resolve in a 100% score", func(t *testing.T) {
		calc := NewCalculator()
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-3", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-4", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-5", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-6", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-7", types.ComplianceStatusPassed))
		assert.Equal(t, 100, int(calc.Score()))
		assert.Equal(t, 7, calc.ControlCount())
		assert.Equal(t, 7, calc.FindingCount())
	})

	t.Run("4 passed, 1 failed, 1 warning, 1 not available should resolve in a 71% score", func(t *testing.T) {
		calc := NewCalculator()
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-3", types.ComplianceStatusFailed))
		calc.ProcessFinding(generateFinding("control-4", types.ComplianceStatusWarning))
		calc.ProcessFinding(generateFinding("control-5", types.ComplianceStatusNotAvailable))
		calc.ProcessFinding(generateFinding("control-6", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-7", types.ComplianceStatusPassed))
		assert.Equal(t, 71, int(calc.Score()))
		assert.Equal(t, 7, calc.ControlCount())
		assert.Equal(t, 7, calc.FindingCount())
	})

	t.Run("2 controls each 2 findings, one has a warning should resolve in 50%", func(t *testing.T) {
		calc := NewCalculator()
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusWarning))
		assert.Equal(t, 50, int(calc.Score()))
		assert.Equal(t, 2, calc.ControlCount())
		assert.Equal(t, 4, calc.FindingCount())
	})

	t.Run("2 controls each 2 findings, one has a warning should resolve in 50% (different order)", func(t *testing.T) {
		calc := NewCalculator()
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusWarning))
		calc.ProcessFinding(generateFinding("control-1", types.ComplianceStatusPassed))
		calc.ProcessFinding(generateFinding("control-2", types.ComplianceStatusPassed))
		assert.Equal(t, 50, int(calc.Score()))
		assert.Equal(t, 2, calc.ControlCount())
		assert.Equal(t, 4, calc.FindingCount())
	})
}
