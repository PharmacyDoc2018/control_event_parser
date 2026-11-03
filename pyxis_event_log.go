package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/PharmacyDoc2018/control_event_parser/database"
	"github.com/google/uuid"
)

const PyxisListFileName = "PyxisList.json"

type PyxisList struct {
	List []string
}

type PyxisEvent struct {
	ItemTransactionKey    uuid.UUID
	UserName              string
	UserID                string
	StorageSpace          string
	ItemID                string
	MedClassCode          string
	MedDisplayName        string
	TransactionType       string
	TxDateTime            time.Time
	EnteredQuantity       float64
	EnteredUOMDisplayCode string
	AmountReferenced      float64
	AmountReferencedUnits string
	BegInventory          float64
	EndInventory          float64
	WitnessName           string
	WitnessID             string
	MRN                   string
}

type PyxisEventLog struct {
	Log               []PyxisEvent
	controlEventLog   *ControlEventLog
	StartDateTime     time.Time
	LastEventDateTime time.Time
	PyxisName         string
}

func ParseEvent(event database.PyxisEventResponse) PyxisEvent {
	pyxisEvent := PyxisEvent{}
	b, _ := event.ItemTransactionKey.MarshalBinary()
	pyxisEvent.ItemTransactionKey = uuid.UUID{
		b[3], b[2], b[1], b[0],
		b[5], b[4],
		b[7], b[6],
		b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15],
	}

	if event.UserName.Valid {
		pyxisEvent.UserName = event.UserName.String
	} else {
		pyxisEvent.UserName = ""
	}

	if event.UserID.Valid {
		pyxisEvent.UserID = event.UserID.String
	} else {
		pyxisEvent.UserID = ""
	}

	if event.StorageSpace.Valid {
		pyxisEvent.StorageSpace = event.StorageSpace.String
	} else {
		pyxisEvent.StorageSpace = ""
	}

	if event.ItemID.Valid {
		pyxisEvent.ItemID = event.ItemID.String
	} else {
		pyxisEvent.ItemID = ""
	}

	if event.MedClassCode.Valid {
		pyxisEvent.MedClassCode = event.MedClassCode.String
	} else {
		pyxisEvent.MedClassCode = ""
	}

	if event.MedDisplayName.Valid {
		pyxisEvent.MedDisplayName = event.MedDisplayName.String
	} else {
		pyxisEvent.MedDisplayName = ""
	}

	if event.TransactionType.Valid {
		pyxisEvent.TransactionType = event.TransactionType.String
	} else {
		pyxisEvent.TransactionType = ""
	}

	if event.TxDateTime.Valid {
		pyxisEvent.TxDateTime = event.TxDateTime.Time
	} else {
		pyxisEvent.TxDateTime = time.Time{}
	}

	if event.EnteredQuantity.Valid {
		pyxisEvent.EnteredQuantity = event.EnteredQuantity.Float64
	} else {
		pyxisEvent.EnteredQuantity = 0.0000
	}

	if event.EnteredUOMDisplayCode.Valid {
		pyxisEvent.EnteredUOMDisplayCode = event.EnteredUOMDisplayCode.String
	} else {
		pyxisEvent.EnteredUOMDisplayCode = ""
	}

	if event.AmountReferenced.Valid {
		pyxisEvent.AmountReferenced = event.AmountReferenced.Float64
	} else {
		pyxisEvent.AmountReferenced = 0.0000
	}

	if event.AmountReferencedUnits.Valid {
		pyxisEvent.AmountReferencedUnits = event.AmountReferencedUnits.String
	} else {
		pyxisEvent.AmountReferencedUnits = ""
	}

	if event.BegInventory.Valid {
		pyxisEvent.BegInventory = event.BegInventory.Float64
	} else {
		pyxisEvent.BegInventory = 0.0000
	}

	if event.EndInventory.Valid {
		pyxisEvent.EndInventory = event.EndInventory.Float64
	} else {
		pyxisEvent.EndInventory = 0.0000
	}

	if event.WitnessName.Valid {
		pyxisEvent.WitnessName = event.WitnessName.String
	} else {
		pyxisEvent.WitnessName = ""
	}

	if event.WitnessID.Valid {
		pyxisEvent.WitnessID = event.WitnessID.String
	} else {
		pyxisEvent.WitnessID = ""
	}

	if event.MRN.Valid {
		pyxisEvent.MRN = event.MRN.String
	} else {
		pyxisEvent.MRN = ""
	}

	return pyxisEvent
}

func (p *PyxisEventLog) checkForNewControlEvents() *logResponder {
	logger := logResponder{}
	logger.AddInfo(fmt.Sprintf("Checking for new control events for %s Pyxis", p.PyxisName))

	//-- Define med class codes that are controlled in map for checking
	controlClassCodes := map[string]struct{}{
		"2": struct{}{},
		"3": struct{}{},
		"4": struct{}{},
		"5": struct{}{},
	}

	//-- Create slice of control pyxis events
	controlEvents := []PyxisEvent{}
	for _, event := range p.Log {
		switch event.TransactionType {
		case "Remove":
			fallthrough
		case "Waste":
			fallthrough
		case "IntWaste":
			fallthrough
		case "Return to bin":
			if _, okay := controlClassCodes[event.MedClassCode]; okay {
				controlEvents = append(controlEvents, event)
			}
		}
	}

	unmatchedEvents := []PyxisEvent{}
	loggedControlEvents := p.controlEventLog.GetLoggedControlEventKeys()

	for _, controlEvent := range controlEvents {
		if _, okay := loggedControlEvents[controlEvent.ItemTransactionKey]; !okay {
			unmatchedEvents = append(unmatchedEvents, controlEvent)
		}
	}

	if len(unmatchedEvents) == 0 {
		logger.AddInfo("No new control events found")
	} else {
		logger.AddInfo(fmt.Sprintf("%d new control events found. Adding to unmatched control log", len(unmatchedEvents)))
	}

	p.controlEventLog.UnmatchedEvents = unmatchedEvents
	return &logger

}

func (p *PyxisList) Add(pyxis string) error {
	for _, device := range p.List {
		if device == pyxis {
			return fmt.Errorf("error. pyxis with that name already exists")
		}
	}

	p.List = append(p.List, pyxis)
	return nil
}

func (p *PyxisList) Remove(pyxis string) error {
	for i, device := range p.List {
		if device == pyxis {
			p.List = append(p.List[:i], p.List[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("error. %s pyxis not found", pyxis)
}

func initPyxisList() *PyxisList {
	p := PyxisList{}
	p.List = []string{}
	return &p
}

func (p *PyxisList) Save(dataPath string) (error, *logResponder) {
	lr := logResponder{}
	lr.AddInfo("Saving DepartmentCoverage")

	file, err := os.OpenFile(filepath.Join(dataPath, PyxisListFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		lr.AddError(fmt.Sprintf("Error. Unable to open %s: %s", ERxItemIdLinksFileName, err.Error()))
		return err, &lr
	}
	defer file.Close()

	data, err := json.Marshal(p)
	if err != nil {
		lr.AddError(fmt.Sprintf("Error marshalling Pyxis List: %s", err.Error()))
		return err, &lr
	}

	_, err = file.Write(data)
	if err != nil {
		lr.AddError(fmt.Sprintf("Error writting marshalled Pyxis List to file %s: %s", departmentCoverageFileName, err.Error()))
		return err, &lr
	}

	return nil, &lr

}

func (p *PyxisList) Load(process *Process) (error, *logResponder) {
	lr := logResponder{}
	lr.AddInfo("Loading DepartmentCoverage")

	//-- Check if save file exists. Create file if not found.
	if _, err := os.Stat(filepath.Join(process.pathToData, PyxisListFileName)); err != nil {
		if os.IsNotExist(err) {
			lr.AddInfo(fmt.Sprintf("%s not found. Attempting to create file", departmentCoverageFileName))
			file, err := os.OpenFile(filepath.Join(process.pathToData, PyxisListFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				lr.AddError(fmt.Sprintf("Error. Unable to create DepartmentCoverage save file: %s", err.Error()))
				return fmt.Errorf("error. unable to create DepartmentCoverage save file: %s", err.Error()), &lr
			}
			_, err = file.WriteString("{}")
			if err != nil {
				lr.AddError(fmt.Sprintf("Error. Unable to write DepartmentCoverage save file: %s", err.Error()))
				return fmt.Errorf("error. unable to write DepartmentCoverage save file: %s", err.Error()), &lr
			}
			file.Close()
		}
	}

	//-- Load json data from save file
	data, err := os.ReadFile(filepath.Join(process.pathToData, PyxisListFileName))
	if err != nil {
		lr.AddError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err, &lr
	}

	//-- Unmarshal json into process memory
	err = json.Unmarshal(data, process.FacilityPyxisList)
	if err != nil {
		lr.AddError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err, &lr

	}

	lr.AddInfo("DepartmentCoverage loaded successfully")
	return nil, &lr

}
