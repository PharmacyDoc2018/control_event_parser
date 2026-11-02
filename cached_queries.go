package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
)

func getPyxisEvents(p *Process, params database.GetPyxisEventsForDeviceByDateRangeParams) ([]database.PyxisEventResponse, error) {
	key := "GetPyxisEventsForDeviceByDateRange"
	key += params.Device
	key += params.Start.Format("2006-01-02 15:04")
	key += params.End.Format("2006-01-02 15:04")

	p.logger.LogInfo(fmt.Sprintf("Query started. Getting Pyxis events for %s from %s to %s", params.Device, params.Start.Format("2006-01-02 1504"), params.End.Format("2006-01-02 1504")))

	data, okay := p.cache.Get(key)
	if okay {
		p.logger.LogInfo("Query results found in cache")
		events := []database.PyxisEventResponse{}
		err := json.Unmarshal(data, &events)
		if err != nil {
			return nil, err
		}
		return events, nil
	}

	p.logger.LogInfo("Query results not found in cache. Querying database")
	events, err := p.dbq.GetPyxisEventsForDeviceByDateRange(context.Background(), params)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error executing SQL query: %s", err.Error()))
		return nil, err
	}

	p.logger.LogInfo("Query successful. Adding to cache")
	data, err = json.Marshal(&events)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error marshalling Pyxis events: %s", err.Error()))
		return nil, err
	}
	p.cache.Add(key, data)

	return events, nil

}

func getMarActions(p *Process, params database.GetMarAdminActionsByPatientsDaysMedIDsParams) ([]database.MarActionResponse, error) {
	if p.state.Mode() == TestMode {
		p.logger.LogInfo("Test mode detected. Func getMarActions returning []database.MarResponse from Process.testMarRes")
		return p.testMarActionRes, nil
	}

	key := "GetMarAdminActionsByPatientDayMedIDs"
	key += params.DateStart.Format("2006-01-02 1504")
	key += params.DateEnd.Format("2006-01-02 1504")

	for _, deptID := range params.DeptIDs {
		key += deptID
	}

	for _, mrn := range params.Mrns {
		key += mrn
	}

	for _, medID := range params.MedIDs {
		key += medID
	}

	p.logger.LogInfo(fmt.Sprintf("Query Started. Getting MAR actions for %d MRN(s) in department %s from %s to %s for %d medID(s)",
		len(params.Mrns),
		params.DeptIDs,
		params.DateStart.Format("2006-01-02"),
		params.DateEnd.Format("2006-01-02"),
		len(params.MedIDs)))

	data, okay := p.cache.Get(key)
	if okay {
		p.logger.LogInfo("Query results found in cache")
		actions := []database.MarActionResponse{}
		err := json.Unmarshal(data, &actions)
		if err != nil {
			return nil, err
		}
		return actions, nil
	}

	p.logger.LogInfo("Query results not found in cache. Querying database")
	actions, err := p.dbq.GetMarAdminActionsByPatientsDaysMedIDs(context.Background(), params)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error executing SQL query: %s", err.Error()))
		return nil, err
	}

	p.logger.LogInfo("Query successful. Adding to cache")
	data, err = json.Marshal(&actions)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error marshalling MAR actions: %s", err.Error()))
		return nil, err
	}
	p.cache.Add(key, data)

	return actions, nil

}

func getMarActionsByOrderNumber(p *Process, orderNumber string) ([]database.MarActionResponse, error) {
	key := "GetMarActionsByOrderNumber"
	key += orderNumber

	p.logger.LogInfo(fmt.Sprintf("Query started. Getting MAR Actions for order number %s", orderNumber))

	data, okay := p.cache.Get(key)
	if okay {
		p.logger.LogInfo("Query results found in cache")
		actions := []database.MarActionResponse{}
		err := json.Unmarshal(data, &actions)
		if err != nil {
			return nil, err
		}
		return actions, nil
	}

	p.logger.LogInfo("Query results not found in cache. Querying database")
	actions, err := p.dbq.GetMarActionsByOrderNumber(context.Background(), orderNumber)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error executing SQL query: %s", err.Error()))
		return nil, err
	}

	p.logger.LogInfo("Query successful. Adding to cache")
	data, err = json.Marshal(&actions)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error marshalling MAR actions: %s", err.Error()))
		return nil, err
	}
	p.cache.Add(key, data)

	return actions, nil
}

func getPyxisEventById(p *Process, id string) (database.PyxisEventResponse, error) {
	key := "GetPyxisEventByItemTransactionKey"
	key += id

	p.logger.LogInfo(fmt.Sprintf("Query started. Getting Pyxis event %s", id))

	data, okay := p.cache.Get(key)
	if okay {
		p.logger.LogInfo("Query results found in cache")
		event := database.PyxisEventResponse{}
		err := json.Unmarshal(data, &event)
		if err != nil {
			return database.PyxisEventResponse{}, err
		}
		return event, nil
	}

	p.logger.LogInfo("Query results not found in cache. Querying database")
	event, err := p.dbq.GetPyxisEventByItemTransactionKey(context.Background(), id)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error executing SQL query: %s", err.Error()))
		return database.PyxisEventResponse{}, err
	}

	p.logger.LogInfo("Query successful. Adding to cache")
	data, err = json.Marshal(&event)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error marshalling Pyxis event: %s", err.Error()))
		return database.PyxisEventResponse{}, err
	}
	p.cache.Add(key, data)

	return event, nil

}
