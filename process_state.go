package main

import (
	"strconv"
)

type processMode int

const (
	Normal = iota
	LocalOnly
	SafetyMode
	TestMode
)

type processState struct {
	mode               processMode
	erxs               bool
	itemIds            bool
	eRxItemIdLinks     bool
	departmentCoverage bool
	dbConnection       bool
	pyxisList          bool
}

func (p *processState) UpdateState() {
	if p.mode == TestMode {
		return
	}

	if !p.erxs {
		p.mode = SafetyMode
		return
	}

	if !p.itemIds {
		p.mode = SafetyMode
		return
	}

	if !p.eRxItemIdLinks {
		p.mode = SafetyMode
		return
	}

	if !p.eRxItemIdLinks {
		p.mode = SafetyMode
		return
	}

	if !p.departmentCoverage {
		p.mode = SafetyMode
	}

	if !p.pyxisList {
		p.mode = SafetyMode
	}

	if !p.dbConnection {
		p.mode = LocalOnly
		return
	}

	p.mode = Normal
}

func (p *processState) Mode() processMode {
	return p.mode
}

func (p *processState) GetState() string {
	res := ""

	res = res + "Mode: " + strconv.Itoa(int(p.mode)) + "\n\n"
	res = res + "Database Connection: " + boolStatus(p.dbConnection) + "\n"
	res = res + "ERXs: " + boolStatus(p.erxs) + "\n"
	res = res + "ItemIDs: " + boolStatus(p.itemIds) + "\n"
	res = res + "ERxItemIdLinks: " + boolStatus(p.eRxItemIdLinks) + "\n"
	res = res + "Department Coverage: " + boolStatus(p.departmentCoverage) + "\n"
	res = res + "Pyxis List: " + boolStatus((p.pyxisList)) + "\n"

	return res
}

func (p *processState) ERXsOkay() bool {
	return p.erxs
}

func (p *processState) ERXsLoadSuccessful() {
	p.erxs = true
	p.UpdateState()
}

func (p *processState) ItemIDsOkay() bool {
	return p.itemIds
}

func (p *processState) ItemIDsLoadSuccessful() {
	p.itemIds = true
	p.UpdateState()
}

func (p *processState) ERxItemIdLinksOkay() bool {
	return p.eRxItemIdLinks
}

func (p *processState) ERxItemIdLinksSuccessful() {
	p.eRxItemIdLinks = true
	p.UpdateState()
}

func (p *processState) DepartmentCoverageOkay() bool {
	return p.departmentCoverage
}

func (p *processState) DepartmentCoverageSuccessful() {
	p.departmentCoverage = true
	p.UpdateState()
}

func (p *processState) PyxisListOkay() bool {
	return p.pyxisList
}

func (p *processState) PyxisListSuccessful() {
	p.pyxisList = true
	p.UpdateState()
}

func (p *processState) DbConnectionOkay() bool {
	return p.dbConnection
}

func (p *processState) DbConnectionSuccessful() {
	p.dbConnection = true
	p.UpdateState()
}

func (p *processState) DbConnectionFail() {
	p.dbConnection = false
	p.UpdateState()
}

func initProcessState() *processState {
	return &processState{
		mode: SafetyMode,
	}
}

func boolStatus(status bool) string {
	if status {
		return "Good"
	} else {
		return "Failed"
	}
}
