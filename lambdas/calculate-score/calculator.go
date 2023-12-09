package main

func NewCalculator() *Calculator {
	history := make(map[Status][]string)
	return &Calculator{
		total:          0,
		failed:         0,
		passed:         0,
		processHistory: history,
	}
}

type Calculator struct {
	total          int
	failed         int
	passed         int
	controls       int
	findings       int
	processHistory map[Status][]string
}

type Status string

const (
	StatusPassed       Status = "PASSED"
	StatusFailed       Status = "FAILED"
	StatusNotProcessed Status = "NOT YET"
)

func (x *Calculator) ProcessFinding(finding *Finding) {
	x.findings++
	status := x.resolveStatus(finding)

	switch x.hasBeenProcessed(finding.GeneratorId) {
	// The control has not been processed yet, so we will increment the current status.
	case StatusNotProcessed:
		if status == StatusPassed {
			x.passed++
		} else {
			x.failed++
		}
		x.processHistory[status] = append(x.processHistory[status], finding.GeneratorId)
		x.total++
	// The control was already been processed as passed.
	case StatusPassed:
		// When a finding was previously marked as passed, we need to revert the passed count and add a failed count.
		if status == StatusFailed {
			x.passed--
			x.failed++
			x.processHistory[status] = append(x.processHistory[status], finding.GeneratorId)
		}
	}
}

func (x *Calculator) Score() float64 {
	if x.total == 0 {
		return float64(100)
	}
	return (float64(x.passed) / float64(x.total)) * 100
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
	return x.controls
}

func (x *Calculator) FindingCount() int {
	return x.findings
}
