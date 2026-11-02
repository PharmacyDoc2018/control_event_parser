package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
	"github.com/gocarina/gocsv"
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

func (p *PyxisEventLog) Save(dataPath string) *logError {
	//-- If old json file exists, remove it
	_, err := os.Stat(filepath.Join(dataPath, pyxisEventLogsFolder, p.PyxisName+".csv"))
	if err == nil {
		os.Remove(filepath.Join(dataPath, pyxisEventLogsFolder, p.PyxisName+".csv"))
	}

	//-- Marshall and write pyxis event log data
	file, err := os.OpenFile(filepath.Join(dataPath, pyxisEventLogsFolder, p.PyxisName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return &logError{
			logMessage: fmt.Sprintf("Error opening %s event log: %s", p.PyxisName, err.Error()),
			errMessage: fmt.Sprintf("error opening %s event log: %s", p.PyxisName, err.Error()),
		}
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(p)
	if err != nil {
		return &logError{
			logMessage: fmt.Sprintf("Error saving %s pyxis event log: %s", p.PyxisName, err.Error()),
			errMessage: fmt.Sprintf("error saving %s pyxis event log: %s", p.PyxisName, err.Error()),
		}
	}

	return nil

}

func (p *PyxisEventLog) Load(dataPath string) *logError {
	//-- Read data from pyxis event log file
	_, err := os.Stat(filepath.Join(dataPath, pyxisEventLogsFolder, p.PyxisName))
	if err != nil { //-- data from json file
		if os.IsNotExist(err) {
			//-- Check for old csv file type
			_, err := os.Stat(filepath.Join(dataPath, pyxisEventLogsFolder, p.PyxisName+".csv"))
			if err != nil {
				if os.IsNotExist(err) {
					//-- Data doesn't exist at all
					return &logError{
						logMessage: fmt.Sprintf("Error. File for %s pyxis event log not found.", p.PyxisName),
						errMessage: fmt.Sprintf("error. file for %s pyxis event log not found.", p.PyxisName),
					}
				} else {
					//-- csv data exists, but file data not accessable
					return &logError{
						logMessage: fmt.Sprintf("Error. Cannot access file for %s pyxis event log: %s.", p.PyxisName, err.Error()),
						errMessage: fmt.Sprintf("error. cannot access file for %s pyxis event log: %s.", p.PyxisName, err.Error()),
					}
				}
			}

			//-- Load data from old csv file --//
			//-- open csv file
			file, err := os.Open(filepath.Join(dataPath, pyxisEventLogsFolder, p.PyxisName+".csv"))
			if err != nil {
				return &logError{
					logMessage: fmt.Sprintf("Error. Unable to load %s Pyxis event log: %s", p.PyxisName, err.Error()),
					errMessage: fmt.Sprintf("error. Unable to load %s pyxis event log: %s", p.PyxisName, err.Error()),
				}
			}
			defer file.Close()

			//-- unmarshal csv into pyxis log
			gocsv.UnmarshalFile(file, &p.Log)

			//-- read settings json file
			data, err := os.ReadFile(filepath.Join(dataPath, pyxisEventLogSettingsFolder, p.PyxisName+".json"))
			if err != nil {
				return &logError{
					logMessage: fmt.Sprintf("Error. Unable to read %s Pyxis event log settings: %s", p.PyxisName, err.Error()),
					errMessage: fmt.Sprintf("error. Unable to read %s pyxis event log settings: %s", p.PyxisName, err.Error()),
				}
			}

			//-- unmarshal json data into struct
			settings := struct {
				StartDateTime     time.Time
				LastEventDateTime time.Time
			}{}
			err = json.Unmarshal(data, &settings)
			if err != nil {
				return &logError{
					logMessage: fmt.Sprintf("Error unmarshalling settings data for %s: %s", p.PyxisName, err.Error()),
					errMessage: fmt.Sprintf("error unmarshalling settings data for %s: %s", p.PyxisName, err.Error()),
				}
			}
			p.StartDateTime = settings.StartDateTime
			p.LastEventDateTime = settings.LastEventDateTime

		}

		//-- Load data from binary file --//
	} else {
		f, err := os.Open(filepath.Join(dataPath, pyxisEventLogsFolder, p.PyxisName))
		if err != nil {
			return &logError{
				logMessage: fmt.Sprintf("Error. Unable to open pyxis event log for %s.", p.PyxisName),
				errMessage: fmt.Sprintf("error. Unable to open pyxis event log for %s.", p.PyxisName),
			}
		}
		defer f.Close()

		decoder := gob.NewDecoder(f)
		err = decoder.Decode(&p)
		if err != nil {
			return &logError{
				logMessage: fmt.Sprintf("Error. Unable to decode binary pyxis event log for %s.", p.PyxisName),
				errMessage: fmt.Sprintf("error. unable to decode binary pyxis event log for %s.", p.PyxisName),
			}
		}
	}

	//-- load control event log
	logErr := p.controlEventLog.Load(dataPath, p)
	if logErr != nil {
		return logErr
	}

	return nil
}

func initPyxisList() *PyxisList {
	p := PyxisList{}
	p.List = []string{}
	return &p
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

func (p *PyxisList) Load() (error, *logResponder) {
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
