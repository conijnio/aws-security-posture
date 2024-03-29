package main

import (
	"log"
	"strings"
)

func NewCalculator(expectedControls []string) *Calculator {
	history := make(map[Status][]string)
	return &Calculator{
		expectedControls: expectedControls,
		total:            0,
		failed:           0,
		passed:           0,
		processHistory:   history,
	}
}

type Calculator struct {
	expectedControls []string
	total            int
	failed           int
	passed           int
	controls         int
	findings         int
	processHistory   map[Status][]string
}

type Status string

const (
	StatusPassed       Status = "PASSED"
	StatusFailed       Status = "FAILED"
	StatusNotProcessed Status = "NOT YET"
)

func (x *Calculator) resolveIdentifier(finding *Finding, groupBy string) string {
	switch groupBy {
	case "Title":
		for _, control := range x.expectedControls {
			if strings.HasPrefix(finding.Title, control) {
				return control
			}
		}

		return finding.Title
	}

	return finding.GeneratorId
}

func (x *Calculator) ProcessFinding(finding *Finding, groupBy string) {
	x.findings++
	status := x.resolveStatus(finding)
	identifier := x.resolveIdentifier(finding, groupBy)
	log.Printf("Resolved identifier: %s\n", identifier)

	switch x.hasBeenProcessed(identifier) {
	// The control has not been processed yet, so we will increment the current status.
	case StatusNotProcessed:
		if status == StatusPassed {
			x.passed++
		} else {
			x.failed++
		}
		x.processHistory[status] = append(x.processHistory[status], identifier)
		x.total++
	// The control was already been processed as passed.
	case StatusPassed:
		// When a finding was previously marked as passed, we need to revert the passed count and add a failed count.
		if status == StatusFailed {
			x.passed--
			x.failed++
			x.processHistory[status] = append(x.processHistory[status], identifier)
		}
	}
}

func (x *Calculator) Score() float64 {
	if x.total == 0 {
		return float64(100)
	}

	return (float64(x.ControlPassedCount()) / float64(x.ControlCount())) * 100
}

func (x *Calculator) hasBeenProcessed(identifier string) Status {
	// When the identifier is already processed as failed we directly consider it processed.
	for _, processedIdentifier := range x.processHistory[StatusFailed] {
		if processedIdentifier == identifier {
			return StatusFailed
		}
	}

	for _, processedIdentifier := range x.processHistory[StatusPassed] {
		if processedIdentifier == identifier {
			return StatusPassed
		}
	}

	x.controls++
	return StatusNotProcessed
}

func (x *Calculator) resolveStatus(finding *Finding) Status {
	switch finding.Status {
	case "FAILED":
		return StatusFailed
	case "WARNING":
		return StatusFailed
	}

	return StatusPassed
}

func (x *Calculator) ControlCount() int {
	if len(x.expectedControls) > 0 {
		x.controls = len(x.expectedControls)
	}
	return x.controls
}

func (x *Calculator) ControlFailedCount() int {
	return x.failed
}

func (x *Calculator) ControlPassedCount() int {
	return x.ControlCount() - x.ControlFailedCount()
}

func (x *Calculator) FindingCount() int {
	return x.findings
}
