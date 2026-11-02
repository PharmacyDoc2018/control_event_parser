package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const departmentCoverageFileName = "DepartmentCoverage.json"

type Department struct {
	ID   string
	Name string
}

type DepartmentCoverage struct {
	List []Department
}

func (d *DepartmentCoverage) GetCoveredDepartments() []Department {
	return d.List
}

func (d *DepartmentCoverage) GetDeptName(deptID string) (string, error) {
	for _, dept := range d.List {
		if dept.ID == deptID {
			return dept.Name, nil
		}
	}

	return "", fmt.Errorf("error. no depatrment with ID %s found", deptID)
}

func (d *DepartmentCoverage) addNewFacilityDepartment(deptID, deptName string) error {
	for _, dept := range d.List {
		if dept.ID == deptID {
			return fmt.Errorf("error. department %s already exists with id %s", dept.ID, dept.Name)
		}
	}

	d.List = append(d.List, Department{deptID, deptName})
	return nil
}

func (d *DepartmentCoverage) removeFacilityDepartment(deptID string) error {
	for i, dept := range d.List {
		if dept.ID == deptID {
			d.List = append(d.List[:i], d.List[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("error. department with id %s not found", deptID)
}

func initDepartmentCoverage() *DepartmentCoverage {
	d := DepartmentCoverage{}
	d.List = []Department{}
	return &d
}

func (p *Process) saveDepartmentCoverage() error {
	p.logger.LogInfo("Saving DepartmentCoverage")

	file, err := os.OpenFile(filepath.Join(p.pathToData, departmentCoverageFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error. Unable to open %s: %s", ERxItemIdLinksFileName, err.Error()))
		return err
	}
	defer file.Close()

	data, err := json.Marshal(p.departmentCoverage)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error marshalling DepartmentCoverage: %s", err.Error()))
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error writting marshalled DepartmentCoverage to file %s: %s", departmentCoverageFileName, err.Error()))
		return err
	}

	return nil

}

func (p *Process) loadDepartmentCoverage() error {
	p.logger.LogInfo("Loading DepartmentCoverage")

	//-- Check if save file exists. Create file if not found.
	if _, err := os.Stat(filepath.Join(p.pathToData, departmentCoverageFileName)); err != nil {
		if os.IsNotExist(err) {
			p.logger.LogInfo(fmt.Sprintf("%s not found. Attempting to create file", departmentCoverageFileName))
			file, err := os.OpenFile(filepath.Join(p.pathToData, departmentCoverageFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				p.logger.LogError(fmt.Sprintf("Error. Unable to create DepartmentCoverage save file: %s", err.Error()))
				return fmt.Errorf("error. unable to create DepartmentCoverage save file: %s", err.Error())
			}
			_, err = file.WriteString("{}")
			if err != nil {
				p.logger.LogError(fmt.Sprintf("Error. Unable to write DepartmentCoverage save file: %s", err.Error()))
				return fmt.Errorf("error. unable to write DepartmentCoverage save file: %s", err.Error())
			}
			file.Close()
		}
	}

	//-- Load json data from save file
	data, err := os.ReadFile(filepath.Join(p.pathToData, departmentCoverageFileName))
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err
	}

	//-- Unmarshal json into process memory
	err = json.Unmarshal(data, p.departmentCoverage)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err

	}

	p.logger.LogInfo("DepartmentCoverage loaded successfully")
	p.state.DepartmentCoverageSuccessful()
	return nil

}
