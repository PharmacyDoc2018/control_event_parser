package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
)

type EventType int

const (
	pyxisEvent = iota
	marAction
	correctionEvent
)

type EventTrailItem struct {
	Type       EventType
	PyxisEvent PyxisEvent
	MarAction  MarAction
}

type EventTrail struct {
	Trail []EventTrailItem
	Valid bool
}

type ControlEventTrail struct {
	MRN         string
	ItemID      string
	Date        time.Time
	EventTrails []EventTrail
	Vaild       bool
}

type ControlEventLog struct {
	Log             []ControlEventTrail
	UnmatchedEvents []PyxisEvent
	pyxisEventLog   *PyxisEventLog
}

func (c *ControlEventLog) Sort() {
	for i := range c.Log {
		for j := range c.Log[i].EventTrails {
			sort.Slice(c.Log[i].EventTrails[j].Trail, func(a, b int) bool {
				dateTimeOne := time.Time{}
				switch c.Log[i].EventTrails[j].Trail[a].Type {
				case pyxisEvent:
					dateTimeOne = c.Log[i].EventTrails[j].Trail[a].PyxisEvent.TxDateTime

				case marAction:
					dateTimeOne = c.Log[i].EventTrails[j].Trail[a].MarAction.SavedTime
				}

				dateTimeTwo := time.Time{}
				switch c.Log[i].EventTrails[j].Trail[b].Type {
				case pyxisEvent:
					dateTimeTwo = c.Log[i].EventTrails[j].Trail[b].PyxisEvent.TxDateTime

				case marAction:
					dateTimeTwo = c.Log[i].EventTrails[j].Trail[b].MarAction.SavedTime
				}

				return dateTimeOne.Before(dateTimeTwo)
			})
		}
	}
	sort.Slice(c.Log, func(i, j int) bool {
		return c.Log[i].Date.Before(c.Log[j].Date)
	})
}

func (c *ControlEventLog) GetLoggedControlEventKeys() map[uuid.UUID]struct{} {
	eventKeys := map[uuid.UUID]struct{}{}

	for _, controlEventTrail := range c.Log {
		for _, eventTrail := range controlEventTrail.EventTrails {
			for _, item := range eventTrail.Trail {
				if item.Type == pyxisEvent {
					eventKeys[item.PyxisEvent.ItemTransactionKey] = struct{}{}
				}
			}
		}
	}

	return eventKeys
}

func (c *ControlEventLog) AddUnmatchedEvents(events []PyxisEvent) {
	c.UnmatchedEvents = append(c.UnmatchedEvents, events...)
}

func (c *ControlEventLog) SortUnmatchedEvents() {
	sort.Slice(c.UnmatchedEvents, func(i, j int) bool {
		return c.UnmatchedEvents[i].TxDateTime.Before(c.UnmatchedEvents[j].TxDateTime)
	})
}

func (c *ControlEventLog) ValidateTrails() {
	for i := range c.Log {
		IsValid := c.Log[i].Vaild
		for _, event := range c.Log[i].EventTrails {
			IsValid = event.Valid
			if !IsValid {
				break
			}
		}
		c.Log[i].Vaild = IsValid
	}
}

func (c *ControlEventLog) Load(dataPath string, pyxisEventLog *PyxisEventLog) *logError {
	//-- Check if control event log file exists. Add if doesn't exist
	if _, err := os.Stat(filepath.Join(dataPath, controlEventLogsFolder, pyxisEventLog.PyxisName)); err != nil {
		if os.IsNotExist(err) {
			if _, err := os.Stat(filepath.Join(dataPath, controlEventLogsFolder, pyxisEventLog.PyxisName+".json")); err != nil {
				if os.IsNotExist(err) {
					pyxisEventLog.controlEventLog = &ControlEventLog{
						Log:             []ControlEventTrail{},
						UnmatchedEvents: []PyxisEvent{},
						pyxisEventLog:   pyxisEventLog,
					}
				} else {
					return &logError{
						logMessage: fmt.Sprintf("Error. Unable to read control event log for %s. Pyxis event log not loaded.", pyxisEventLog.PyxisName),
						errMessage: fmt.Sprintf("error. unable to read control event log for %s. pyxis event log not loaded.", pyxisEventLog.PyxisName),
					}
				}
			}
		}
	}

	//-- Read data from control event log file
	_, err := os.Stat(filepath.Join(dataPath, controlEventLogsFolder, pyxisEventLog.PyxisName))
	if err != nil { //-- data from json file
		if os.IsNotExist(err) {
			//-- Check for old JSON file type
			_, err := os.Stat(filepath.Join(dataPath, controlEventLogsFolder, pyxisEventLog.PyxisName+".json"))
			if err != nil {
				if os.IsNotExist(err) {
					//-- Data doesn't exist at all
					return &logError{
						logMessage: fmt.Sprintf("Error. File for %s control event log not found. Pyxis event log not loaded", pyxisEventLog.PyxisName),
						errMessage: fmt.Sprintf("error. file for %s control event log not found. pyxis event log not loaded", pyxisEventLog.PyxisName),
					}
				} else {
					//-- json data exists, but file data not accessable
					return &logError{
						logMessage: fmt.Sprintf("Error. Cannot access file information for %s control event log: %s. Pyxis event log not loaded", pyxisEventLog.PyxisName, err.Error()),
						errMessage: fmt.Sprintf("error. cannot access file information for %s control event log: %s. Pyxis event log not loaded", pyxisEventLog.PyxisName, err.Error()),
					}
				}
			}

			//-- Old json file found
			f, err := os.Open(filepath.Join(dataPath, controlEventLogsFolder, pyxisEventLog.PyxisName+".json"))
			if err != nil {
				return &logError{
					logMessage: fmt.Sprintf("Error. Unable to open control event log for %s. Pyxis event log not loaded", pyxisEventLog.PyxisName),
					errMessage: fmt.Sprintf("error. Unable to open control event log for %s. pyxis event log not loaded", pyxisEventLog.PyxisName),
				}
			}
			defer f.Close()

			decoder := json.NewDecoder(f)
			err = decoder.Decode(&pyxisEventLog.controlEventLog)
			if err != nil {
				return &logError{
					logMessage: fmt.Sprintf("Error. Unable to decode json control event log for %s. Pyxis event log not loaded.", pyxisEventLog.PyxisName),
					errMessage: fmt.Sprintf("error. Unable to decode json control event log for %s. pyxis event log not loaded.", pyxisEventLog.PyxisName),
				}
			}
		}
	} else { //-- data from binary file
		//-- Load binary file
		f, err := os.Open(filepath.Join(dataPath, controlEventLogsFolder, pyxisEventLog.PyxisName))
		if err != nil {
			return &logError{
				logMessage: fmt.Sprintf("Error. Unable to open control event log for %s. Pyxis event log not loaded", pyxisEventLog.PyxisName),
				errMessage: fmt.Sprintf("error. Unable to open control event log for %s. pyxis event log not loaded", pyxisEventLog.PyxisName),
			}
		}
		defer f.Close()

		decoder := gob.NewDecoder(f)
		err = decoder.Decode(&pyxisEventLog.controlEventLog)
		if err != nil {
			return &logError{
				logMessage: fmt.Sprintf("Error. Unable to decode binary control event log for %s. Pyxis event log not loaded.", pyxisEventLog.PyxisName),
				errMessage: fmt.Sprintf("error. unable to decode binary control event log for %s. pyxis event log not loaded.", pyxisEventLog.PyxisName),
			}
		}
	}

	//-- Link PyxisEventLog pointer in ControlEventLog
	pyxisEventLog.controlEventLog.pyxisEventLog = pyxisEventLog
	return nil
}

func (c *ControlEventLog) Save(p *Process) error {
	//-- If old json file exists, remove it
	_, err := os.Stat(filepath.Join(p.pathToData, controlEventLogsFolder, c.pyxisEventLog.PyxisName+".json"))
	if err == nil {
		os.Remove(filepath.Join(p.pathToData, controlEventLogsFolder, c.pyxisEventLog.PyxisName+".json"))
	}

	//-- Marshall and write control event log data
	controlFile, err := os.OpenFile(filepath.Join(p.pathToData, controlEventLogsFolder, c.pyxisEventLog.PyxisName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error opening %s control event log: %s", c.pyxisEventLog.PyxisName, err.Error()))
		return err
	}
	defer controlFile.Close()

	encoder := gob.NewEncoder(controlFile)
	err = encoder.Encode(c)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error saving %s control event log: %s", c.pyxisEventLog.PyxisName, err.Error()))
		return err
	}

	return nil
}

func (c *ControlEventLog) MatchEvents(pyxisEvents []PyxisEvent, marActions []MarAction, date time.Time, mrn, itemID string) []PyxisEvent {
	allEvents := []EventTrailItem{}

	//---------------- FOR TESTING DELETE LATER ----------------------//
	//fmt.Println("Pyxis events to match:")
	//for _, event := range pyxisEvents {
	//	fmt.Printf("%s event %s on %s by %s for mrn %s\n",
	//		event.MedDisplayName,
	//		event.TransactionType,
	//		event.TxDateTime.Format("2006-01-02 1504"),
	//		event.UserName,
	//		event.MRN)
	//}
	//fmt.Println()

	//fmt.Println("MAR actions to match:")
	//for _, action := range marActions {
	//fmt.Printf("%s action %s on %s by %s for mrn %s\n",
	//action.DisplayName,
	//action.MarAction,
	//action.SavedTime.Format("2006-01-02 1504"),
	//action.UserName,
	//action.MRN)
	//}
	//fmt.Println()
	//-------------------------------------------------------------//

	for _, event := range pyxisEvents {
		allEvents = append(allEvents, EventTrailItem{
			Type:       pyxisEvent,
			PyxisEvent: event,
		})
	}

	for _, action := range marActions {
		allEvents = append(allEvents, EventTrailItem{
			Type:      marAction,
			MarAction: action,
		})
	}

	sort.Slice(allEvents, func(i, j int) bool {
		timeOne := time.Time{}
		timeTwo := time.Time{}

		if allEvents[i].Type == pyxisEvent {
			timeOne = allEvents[i].PyxisEvent.TxDateTime
		} else {
			timeOne = allEvents[i].MarAction.SavedTime
		}

		if allEvents[j].Type == pyxisEvent {
			timeTwo = allEvents[j].PyxisEvent.TxDateTime
		} else {
			timeTwo = allEvents[j].MarAction.SavedTime
		}

		return timeOne.Before(timeTwo)
	})

	//------------ FOR TESTING DELETE LATER --------------------//

	//fmt.Println("Pyxis events and Mar actions sorted into single list:")
	//for i, event := range allEvents {
	//switch event.Type {
	//case pyxisEvent:
	//fmt.Printf("%d. Pyxis Event: %s event %s on %s by %s for mrn %s\n",
	//i+1,
	//event.PyxisEvent.MedDisplayName,
	//event.PyxisEvent.TransactionType,
	//event.PyxisEvent.TxDateTime.Format("2006-01-02 1504"),
	//event.PyxisEvent.UserName,
	//event.PyxisEvent.MRN)

	//case marAction:
	//fmt.Printf("%d. MAR Action: %s action %s on %s by %s for mrn %s\n",
	//i+1,
	//event.MarAction.DisplayName,
	//event.MarAction.MarAction,
	//event.MarAction.SavedTime.Format("2006-01-02 1504"),
	//event.MarAction.UserName,
	//event.MarAction.MRN)

	//}
	//}
	//fmt.Println()
	//-------------------------------------------------------//

	controlEventTrail := ControlEventTrail{
		MRN:         mrn,
		ItemID:      itemID,
		Date:        date,
		EventTrails: []EventTrail{},
		Vaild:       false,
	}

	trackedTransactions := map[string]struct{}{
		"IntWaste":         struct{}{},
		"Remove":           struct{}{},
		"Remove Cancelled": struct{}{},
		"Return to bin":    struct{}{},
		"Waste":            struct{}{},
	}

	matchStatus := struct {
		InitialRemoveIndex     int
		InitialRemoveAmount    float64
		SubsequentRemoveIndexs []int
		CurrentAmount          float64
		Done                   bool
	}{
		InitialRemoveIndex:     0,
		InitialRemoveAmount:    0.0,
		SubsequentRemoveIndexs: []int{},
		CurrentAmount:          0.0,
		Done:                   false,
	}

	//fmt.Printf("Starting matching loop. MatchStatus.Done = %t\n\n", matchStatus.Done) //--DELETE AFTER TEST
	for !matchStatus.Done {
		//-- Reset Match Status
		matchStatus.InitialRemoveIndex = 0
		matchStatus.InitialRemoveAmount = 0.0
		matchStatus.SubsequentRemoveIndexs = []int{}
		matchStatus.CurrentAmount = 0.0
		fmt.Println("Match Status Reset!") //-- DELETE AFTER TEST
		for i, event := range allEvents {
			//------------------------------DELETE AFTER TEST----------------------------//
			//fmt.Println("Starting evaluation of next event. Current Match Status:")
			//fmt.Printf("Initial Remove Index: %d\n", matchStatus.InitialRemoveIndex)
			//fmt.Printf("Initial Remove Amount: %f\n", matchStatus.InitialRemoveAmount)
			//fmt.Print("Subsequent Remove Indexes: ")
			//for _, n := range matchStatus.SubsequentRemoveIndexs {
			//	fmt.Printf("%d, ", n)
			//}
			//fmt.Println()
			//fmt.Printf("Current amount: %f\n", matchStatus.CurrentAmount)
			//fmt.Printf("Is Done: %t\n", matchStatus.Done)
			//fmt.Println()
			//---------------------------------------------------------------------------//
			switch event.Type {
			case pyxisEvent:
				switch event.PyxisEvent.TransactionType {
				case "Remove":
					//------------ DELETE AFTER TEST ____________________________________//
					//fmt.Println("Current event is a pyxis remove event")
					//fmt.Printf("%s: %s by %s on %s\n",
					//event.PyxisEvent.TransactionType,
					//event.PyxisEvent.MedDisplayName,
					//event.PyxisEvent.UserName,
					//event.PyxisEvent.TxDateTime.Format("2006-01-02 1504"))
					//-------------------------------------------------------------------//
					if matchStatus.InitialRemoveAmount == 0.0 {
						//fmt.Println("This remove is an initial remove event") //-- DELETE AFTER TEST
						matchStatus.InitialRemoveAmount = event.PyxisEvent.AmountReferenced
						matchStatus.InitialRemoveIndex = i
						matchStatus.CurrentAmount = event.PyxisEvent.AmountReferenced
						//-------------DELETE AFTER TEST ----------------------------//
						//fmt.Println("Updated Match Status after remove:")
						//fmt.Printf("Initial Remove Index: %d\n", matchStatus.InitialRemoveIndex)
						//fmt.Printf("Initial Remove Amount: %f\n", matchStatus.InitialRemoveAmount)
						//fmt.Print("Subsequent Remove Indexes: ")
						//for _, n := range matchStatus.SubsequentRemoveIndexs {
						//	fmt.Printf("%d, ", n)
						//}
						//fmt.Println()
						//fmt.Printf("Current amount: %f\n", matchStatus.CurrentAmount)
						//fmt.Printf("Is Done: %t\n", matchStatus.Done)
						//fmt.Println()
						//----------------------------------------------------------------//
					} else {
						//fmt.Println("This is a subsequent remove event") //-- DELETE AFTER TEST
						matchStatus.CurrentAmount = addFloat(matchStatus.CurrentAmount, event.PyxisEvent.AmountReferenced)
						matchStatus.SubsequentRemoveIndexs = append(matchStatus.SubsequentRemoveIndexs, i)
						//-------------DELETE AFTER TEST ----------------------------//
						//fmt.Println("Updated Match Status:")
						//fmt.Printf("Initial Remove Index: %d\n", matchStatus.InitialRemoveIndex)
						//fmt.Printf("Initial Remove Amount: %f\n", matchStatus.InitialRemoveAmount)
						//fmt.Print("Subsequent Remove Indexes: ")
						//for _, n := range matchStatus.SubsequentRemoveIndexs {
						//	fmt.Printf("%d, ", n)
						//}
						//fmt.Println()
						//fmt.Printf("Current amount: %f\n", matchStatus.CurrentAmount)
						//fmt.Printf("Is Done: %t\n", matchStatus.Done)
						//fmt.Println()
						//----------------------------------------------------------------//

					}
				case "Waste":
					//-------------DELETE AFTER TEST ----------------------------//
					//fmt.Println("Current event is a pyxis waste event")
					//fmt.Printf("%s: %s by %s on %s\n",
					//	event.PyxisEvent.TransactionType,
					//	event.PyxisEvent.MedDisplayName,
					//	event.PyxisEvent.UserName,
					//	event.PyxisEvent.TxDateTime.Format("2006-01-02 1504"))
					//----------------------------------------------------------------//

					matchStatus.CurrentAmount = subtractFloat(matchStatus.CurrentAmount, event.PyxisEvent.AmountReferenced)

					//-------------DELETE AFTER TEST ----------------------------//
					//fmt.Println("Updated Match Status:")
					//fmt.Printf("Initial Remove Index: %d\n", matchStatus.InitialRemoveIndex)
					//fmt.Printf("Initial Remove Amount: %f\n", matchStatus.InitialRemoveAmount)
					//fmt.Print("Subsequent Remove Indexes: ")
					//for _, n := range matchStatus.SubsequentRemoveIndexs {
					//	fmt.Printf("%d, ", n)
					//}
					//fmt.Println()
					//fmt.Printf("Current amount: %f\n", matchStatus.CurrentAmount)
					//fmt.Printf("Is Done: %t\n", matchStatus.Done)
					//fmt.Println()
					//----------------------------------------------------------------//

				case "IntWaste":
					matchStatus.CurrentAmount = subtractFloat(matchStatus.CurrentAmount, event.PyxisEvent.AmountReferenced)

				case "Return to bin":
					matchStatus.CurrentAmount = subtractFloat(matchStatus.CurrentAmount, event.PyxisEvent.AmountReferenced)
				}

			case marAction:
				//------------ DELETE AFTER TEST ____________________________________//
				//fmt.Println("Current event is a MAR action event")
				//fmt.Printf("%s: %s by %s on %s\n",
				//	event.MarAction.MarAction,
				//	event.MarAction.DisplayName,
				//	event.MarAction.UserName,
				//	event.MarAction.SavedTime.Format("2006-01-02 1504"))
				//-------------------------------------------------------------------//

				matchStatus.CurrentAmount = subtractFloat(matchStatus.CurrentAmount, event.MarAction.Dose)

				//-------------DELETE AFTER TEST ----------------------------//
				//fmt.Println("Updated Match Status:")
				//fmt.Printf("Initial Remove Index: %d\n", matchStatus.InitialRemoveIndex)
				//fmt.Printf("Initial Remove Amount: %f\n", matchStatus.InitialRemoveAmount)
				//fmt.Print("Subsequent Remove Indexes: ")
				//for _, n := range matchStatus.SubsequentRemoveIndexs {
				//	fmt.Printf("%d, ", n)
				//}
				//fmt.Println()
				//fmt.Printf("Current amount: %f\n", matchStatus.CurrentAmount)
				//fmt.Printf("Is Done: %t\n", matchStatus.Done)
				//fmt.Println()
				//----------------------------------------------------------------//
			}

			//-- Check for negative amount. Return remaining events as unmatched events
			if matchStatus.CurrentAmount < 0.0 {
				unmatchedEvents := []PyxisEvent{}
				for _, event := range allEvents {
					if event.Type == pyxisEvent {
						unmatchedEvents = append(unmatchedEvents, event.PyxisEvent)
					}
				}
				matchStatus.Done = true
				break
			}

			//-- Check for trail end:
			//-- If current amount = 0 and there is at least one remove event -> end trail
			if matchStatus.CurrentAmount == 0.0 && matchStatus.InitialRemoveAmount > 0.0 {
				newTrail := allEvents[0 : i+1]
				if len(newTrail) == len(allEvents) {
					allEvents = []EventTrailItem{}
					matchStatus.Done = true
				} else {
					allEvents = allEvents[i+1:]
				}

				tempTrail := newTrail
				newTrail = []EventTrailItem{}
				for _, item := range tempTrail {
					switch item.Type {
					case pyxisEvent:
						if _, okay := trackedTransactions[item.PyxisEvent.TransactionType]; okay {
							newTrail = append(newTrail, item)
						}

					case marAction:
						newTrail = append(newTrail, item)
					}
				}

				controlEventTrail.EventTrails = append(controlEventTrail.EventTrails, EventTrail{
					Trail: newTrail,
					Valid: true,
				})
				break
			}

			//-- if current amount == initial remove amount AND there are more than one remove events -> piece trail together
			if matchStatus.CurrentAmount == matchStatus.InitialRemoveAmount && len(matchStatus.SubsequentRemoveIndexs) > 0 {
				removeIndexsToRemove := map[int]struct{}{}
				amountSoFar := 0.0
				for j, item := range allEvents[0 : i+1] {
					if item.PyxisEvent.TransactionType == "Remove" &&
						addFloat(item.PyxisEvent.AmountReferenced, amountSoFar) <= matchStatus.CurrentAmount {
						amountSoFar = addFloat(amountSoFar, item.PyxisEvent.AmountReferenced)
						removeIndexsToRemove[j] = struct{}{}
					}
				}

				tempTrail := allEvents[0 : i+1]
				newTrail := []EventTrailItem{}
				addBackTrail := []EventTrailItem{}

				for j, item := range tempTrail {
					switch item.Type {
					case pyxisEvent:
						if _, okay := removeIndexsToRemove[j]; okay {
							addBackTrail = append(addBackTrail, item)
							continue
						}
						if _, okay := trackedTransactions[item.PyxisEvent.TransactionType]; okay {
							newTrail = append(newTrail, item)
						}

					case marAction:
						newTrail = append(newTrail, item)
					}
				}

				controlEventTrail.EventTrails = append(controlEventTrail.EventTrails, EventTrail{
					Trail: newTrail,
					Valid: true,
				})

				if len(tempTrail) == len(allEvents) && len(addBackTrail) == 0 {
					allEvents = []EventTrailItem{}
					matchStatus.Done = true
				} else {
					allEvents = allEvents[i+1:]
					allEvents = append(addBackTrail, allEvents...)
				}
				break
			}

			if i+1 >= len(allEvents) {
				matchStatus.Done = true
			}
		}

	}
	if len(controlEventTrail.EventTrails) > 0 {
		type compareStruct struct {
			MRN    string
			ItemID string
			Date   time.Time
		}

		found := false
		for i, cet := range c.Log {
			newCetComp := compareStruct{
				MRN:    controlEventTrail.MRN,
				ItemID: controlEventTrail.ItemID,
				Date:   controlEventTrail.Date,
			}

			cetComp := compareStruct{
				MRN:    cet.MRN,
				ItemID: cet.ItemID,
				Date:   cet.Date,
			}

			if newCetComp == cetComp {
				found = true
				c.Log[i].EventTrails = append(c.Log[i].EventTrails, controlEventTrail.EventTrails...)
				break
			}
		}
		if !found {
			c.Log = append(c.Log, controlEventTrail)
		}
	}
	unmatchedEvents := []PyxisEvent{}
	for _, event := range allEvents {
		if event.Type == pyxisEvent {
			unmatchedEvents = append(unmatchedEvents, event.PyxisEvent)
		}
	}
	return unmatchedEvents

}

func (c *ControlEventLog) GenerateTrailSlices() [][]EventTrail {
	c.Sort()

	slices := [][]EventTrail{}
	for _, event := range c.Log {
		slices = append(slices, event.EventTrails)
	}

	return slices
}

func (c *ControlEventLog) LinkEventActions(mrn, itemID string, date time.Time, items ...EventTrailItem) *logError {
	if len(items) == 0 {
		return &logError{
			errMessage: "error. no event actions given to link",
			logMessage: "Error. No event actions given to link",
		}

	}

	h, m, s := date.Clock()
	if h != 0 || m != 0 || s != 0 {
		return &logError{
			errMessage: "error. time must be set to midnight of the chosen day",
			logMessage: "Error. Time must be set to midnight of the chosen day",
		}
	}

	tempUnmatchedEvents := []PyxisEvent{}
	NetAmount := 0.0

	for _, item := range items {
		switch item.Type {
		case pyxisEvent:
			//-- Check to make sure event is listed in the UnmatchedEvents
			for i, unmatchedEvent := range c.UnmatchedEvents {
				if item.PyxisEvent.ItemTransactionKey == unmatchedEvent.ItemTransactionKey {
					tempUnmatchedEvents = append(tempUnmatchedEvents, item.PyxisEvent)
					c.UnmatchedEvents = append(c.UnmatchedEvents[:i], c.UnmatchedEvents[i+1:]...)
					break
				}
			}

			//-- Add to net amount for zero check
			switch item.PyxisEvent.TransactionType {
			case "Remove", "Destock", "Empty return bin", "Outdate", "Unload":
				addFloat(NetAmount, item.PyxisEvent.AmountReferenced)

			case "Waste", "IntWaste", "Refill", "Load", "Return to bin", "Return to stock":
				subtractFloat(NetAmount, item.PyxisEvent.AmountReferenced)
			}

		case marAction:
			subtractFloat(NetAmount, item.MarAction.Dose)

		case correctionEvent:
			subtractFloat(NetAmount, item.CorrectionEvent.Amount)
		}
	}

	if NetAmount != 0.0 {
		return &logError{
			errMessage: "error. net amount used must be zero",
			logMessage: "Error. Net amount used must be zero",
		}
	}

	sort.Slice(items, func(i, j int) bool {
		dateOne := time.Time{}
		switch items[i].Type {
		case pyxisEvent:
			dateOne = items[i].PyxisEvent.TxDateTime

		case marAction:
			dateOne = items[i].MarAction.SavedTime
		}

		dateTwo := time.Time{}
		switch items[j].Type {
		case pyxisEvent:
			dateOne = items[j].PyxisEvent.TxDateTime

		case marAction:
			dateOne = items[j].MarAction.SavedTime
		}

		return dateOne.Before(dateTwo)
	})

	eventTrail := EventTrail{
		Trail: items,
		Valid: true,
	}

	found := false
	for i, controlEventTrail := range c.Log {
		if date.Equal(controlEventTrail.Date) && mrn == controlEventTrail.MRN {
			found = true
			c.Log[i].EventTrails = append(c.Log[i].EventTrails, eventTrail)
			sort.Slice(c.Log[i].EventTrails, func(a, b int) bool {
				dateOne := time.Time{}
				switch c.Log[i].EventTrails[a].Trail[0].Type {
				case pyxisEvent:
					dateOne = c.Log[i].EventTrails[a].Trail[0].PyxisEvent.TxDateTime

				case marAction:
					dateOne = c.Log[i].EventTrails[a].Trail[0].MarAction.SavedTime
				}

				dateTwo := time.Time{}
				switch c.Log[i].EventTrails[b].Trail[0].Type {
				case pyxisEvent:
					dateTwo = c.Log[i].EventTrails[b].Trail[0].PyxisEvent.TxDateTime

				case marAction:
					dateTwo = c.Log[i].EventTrails[b].Trail[0].MarAction.SavedTime
				}

				return dateOne.Before(dateTwo)
			})
			break
		}

		if !found {
			c.Log = append(c.Log, ControlEventTrail{
				MRN:    mrn,
				ItemID: itemID,
				Date:   date,
				EventTrails: []EventTrail{
					{
						Trail: items,
						Valid: true,
					},
				},
				Vaild: true,
			})

			sort.Slice(c.Log, func(i, j int) bool {
				return c.Log[i].Date.Before(c.Log[j].Date)
			})
		}
	}

	return nil
}
