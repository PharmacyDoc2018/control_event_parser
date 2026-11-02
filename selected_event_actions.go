package main

import (
	"fmt"
	"sort"
	"strconv"
	"time"
)

type SelectedEventActions struct {
	Map map[EventTrailItem]struct{}
}

func (s *SelectedEventActions) RemoveTrailAndSelect(controlEventLog *ControlEventLog, controlEventTrail *ControlEventTrail, index int) *logError {
	//-- Check index is valid
	if index >= len(controlEventTrail.EventTrails) {
		return &logError{
			errMessage: fmt.Sprintf("error. invalid index: index = %d last index = %d", index, len(controlEventTrail.EventTrails)-1),
			logMessage: fmt.Sprintf("Error. Invalid index: index = %d last index = %d", index, len(controlEventTrail.EventTrails)-1),
		}
	}
	//-- Verify that items aren't already selected
	for _, item := range controlEventTrail.EventTrails[index].Trail {
		if _, okay := s.Map[item]; okay {
			switch item.Type {
			case pyxisEvent:
				return &logError{
					errMessage: fmt.Sprintf("error. pyxis event %s already selected", item.PyxisEvent.ItemTransactionKey.String()),
					logMessage: fmt.Sprintf("Error. Pyxis event %s already selected", item.PyxisEvent.ItemTransactionKey.String()),
				}

			case marAction:
				return &logError{
					errMessage: fmt.Sprintf("error. mar action with order number %s already selected", item.MarAction.OrderNumber),
					logMessage: fmt.Sprintf("Error. MAR action with order number %s already selected", item.MarAction.OrderNumber),
				}
			}
		}
	}

	//-- Add EventTrailItems to the selected map and Pyxis events to unmatched events
	for _, item := range controlEventTrail.EventTrails[index].Trail {
		s.Map[item] = struct{}{}

		if item.Type == pyxisEvent {
			controlEventLog.UnmatchedEvents = append(controlEventLog.UnmatchedEvents, item.PyxisEvent)
		}
	}

	//-- Remove the EventTrail containing selected events from the controlEventTrail
	controlEventTrail.EventTrails = append(controlEventTrail.EventTrails[:index], controlEventTrail.EventTrails[index+1:]...)

	return nil

}

func (s *SelectedEventActions) SelectUnmatchedEvent(controlEventLog *ControlEventLog, index int) *logError {
	//-- Check that index is valid
	if index >= len(controlEventLog.UnmatchedEvents) {
		return &logError{
			errMessage: fmt.Sprintf("error. invalid index: index = %d last index = %d", index, len(controlEventLog.GetLoggedControlEventKeys())),
			logMessage: fmt.Sprintf("Error. Invalid index: index = %d last index = %d", index, len(controlEventLog.GetLoggedControlEventKeys())),
		}
	}

	unmatchedEventTrailItem := EventTrailItem{
		Type:       pyxisEvent,
		PyxisEvent: controlEventLog.UnmatchedEvents[index],
	}

	//-- Verify that event isn't already selected
	if _, okay := s.Map[unmatchedEventTrailItem]; okay {
		return &logError{
			errMessage: fmt.Sprintf("error. pyxis event %s is already selected", unmatchedEventTrailItem.PyxisEvent.ItemTransactionKey.String()),
			logMessage: fmt.Sprintf("Error. Pyxis event %s is already selected", unmatchedEventTrailItem.PyxisEvent.ItemTransactionKey.String()),
		}
	}

	//-- Add unmatchedEvent to selecte map as an EventTrailItem
	s.Map[unmatchedEventTrailItem] = struct{}{}

	return nil

}

func (s *SelectedEventActions) PrintSelected() {
	items := []EventTrailItem{}
	for item := range s.Map {
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		dateTimeOne := time.Time{}
		switch items[i].Type {
		case pyxisEvent:
			dateTimeOne = items[i].PyxisEvent.TxDateTime

		case marAction:
			dateTimeOne = items[i].MarAction.SavedTime
		}

		dateTimeTwo := time.Time{}
		switch items[j].Type {
		case pyxisEvent:
			dateTimeOne = items[j].PyxisEvent.TxDateTime

		case marAction:
			dateTimeOne = items[j].MarAction.SavedTime
		}

		return dateTimeOne.Before(dateTimeTwo)
	})

	for _, item := range items {
		switch item.Type {
		case pyxisEvent:
			printfln("Transaction Type: %s", item.PyxisEvent.TransactionType)
			printfln("Tx Date Time: %s", item.PyxisEvent.TxDateTime.Format("2006-01-02 1504"))
			printfln("User ID: %s", item.PyxisEvent.UserID)
			printfln("User Name: %s", item.PyxisEvent.UserName)
			printfln("Display Name: %s %s", item.PyxisEvent.MedDisplayName)
			printfln("Amount: %s", strconv.FormatFloat(item.PyxisEvent.AmountReferenced, 'f', -1, 64), item.PyxisEvent.AmountReferencedUnits)
			printfln("MRN: %s", item.PyxisEvent.MRN)
			printfln("Witness: %s", item.PyxisEvent.WitnessName)
			fmt.Println()

		case marAction:
			printfln("MAR Action: %s", item.MarAction.MarAction)
			printfln("Saved Name: %s", item.MarAction.SavedTime.Format("2006-01-02 1504"))
			printfln("User ID: %s", item.MarAction.UserID)
			printfln("User Name: %s", item.MarAction.UserName)
			printfln("Display Name: %s %s", item.MarAction.DisplayName)
			printfln("Dose: %s", strconv.FormatFloat(item.MarAction.Dose, 'f', -1, 64), item.MarAction.DoseUnitDescription)
			printfln("MRN: %s", item.MarAction.MRN)
			printfln("Patient Name: %s", item.MarAction.PtName)
			fmt.Println()
		}
	}
}

func (s *SelectedEventActions) Add(item EventTrailItem) *logError {
	if _, okay := s.Map[item]; okay {
		switch item.Type {
		case pyxisEvent:
			id := item.PyxisEvent.ItemTransactionKey.String()
			return &logError{
				logMessage: fmt.Sprintf("Error. Pyxis event %s is already selected", id),
				errMessage: fmt.Sprintf("error. pyxis event %s is already selected", id),
			}

		case marAction:
			orderNumber := item.MarAction.OrderNumber
			return &logError{
				logMessage: fmt.Sprintf("Error. MAR action with order number %s is already selected", orderNumber),
				errMessage: fmt.Sprintf("error. mar action with order number %s is already selected", orderNumber),
			}

		case correctionEvent:
			return &logError{
				logMessage: "Error. That correction event is already selected",
				errMessage: "error. that correction event is already selected",
			}
		}
	}

	s.Map[item] = struct{}{}
	return nil
}
