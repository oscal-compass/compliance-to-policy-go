/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package policy

import (
	"time"
)

// Result represents the kind of result statuses.
type Result uint

const (
	ResultInvalid Result = iota
	ResultFail
	ResultError
	ResultPass
	ResultWarning
)

// String prints a string representation of the result
func (r Result) String() string {
	switch r {
	case ResultInvalid:
		return "INVALID"
	case ResultFail:
		return "fail"
	case ResultError:
		return "error"
	case ResultPass:
		return "pass"
	case ResultWarning:
		return "warning"
	default:
		panic("invalid result")
	}
}

type Property struct {
	Name  string
	Value string
}

type Link struct {
	Description string
	Href        string
}

type Subject struct {
	Title       string
	Type        string
	ResourceID  string
	Result      Result
	EvaluatedOn time.Time
	Reason      string
	Props       []Property
}

type ObservationByCheck struct {
	Title             string
	Description       string
	CheckID           string
	Methods           []string
	Subjects          []Subject
	Collected         time.Time
	RelevantEvidences []Link
	Props             []Property
}

type PVPResult struct {
	ObservationsByCheck []ObservationByCheck
	Links               []Link
}
