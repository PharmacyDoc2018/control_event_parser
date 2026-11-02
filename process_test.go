package main

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
	"github.com/google/uuid"
)

func TestPyxisEventLog(t *testing.T) {
	startTime, err := time.Parse("01/02/2006", "01/01/2026")
	if err != nil {
		t.Fatalf("Failed to make start time: %s", err.Error())
	}

	firstEventTime, _ := time.Parse("01/02/2006 15:04", "01/02/2026 09:48")
	secondEventTime, _ := time.Parse("01/02/2006 15:04", "01/04/2026 13:23")
	thirdEventTime, _ := time.Parse("01/02/2006 15:04", "01/06/2026 10:56")

	p := initProcess()
	p.PyxisEventLogs = []*PyxisEventLog{}

	err = p.db.Ping()
	if err != nil {
		p.dbConnection = false
	} else {
		p.dbConnection = true
	}
	if p.dbConnection {
		p.db.Close()
	}

	err = p.createNewPyxisEventLog("testPyxis", startTime)
	if err != nil {
		t.Errorf("error creating new Pyxis event log: %s", err.Error())
	}
	p.logger.Log(p.PyxisEventLogs[0].AddPyxisEvents([]PyxisEvent{
		{
			ItemTransactionKey:    uuid.New(),
			UserName:              "Testnurse, One",
			UserID:                "abc1a",
			StorageSpace:          "TESTPYXIS_MAIN Drw 4.1-Pkt E4",
			ItemID:                "07571",
			MedClassCode:          "2",
			MedDisplayName:        "oxyCODONE 5 mg TABLET UD",
			TransactionType:       "Count inventory",
			TxDateTime:            firstEventTime,
			EnteredQuantity:       96.0000,
			EnteredUOMDisplayCode: "Dosage Form",
			AmountReferenced:      480.0000,
			AmountReferencedUnits: "mg",
			BegInventory:          96.0000,
			EndInventory:          96.0000,
			WitnessName:           "Testnurse, Two",
			WitnessID:             "abc2b",
			MRN:                   "",
		},
		{
			ItemTransactionKey:    uuid.New(),
			UserName:              "Testnurse, Two",
			UserID:                "abc2b",
			StorageSpace:          "TESTPYXIS_MAIN Drw 3.1-Pkt D3",
			ItemID:                "06896",
			MedClassCode:          "Non-Controlled",
			MedDisplayName:        "PROCHLORPERAZINE 10 mg TABLET UD",
			TransactionType:       "Remove",
			TxDateTime:            secondEventTime,
			EnteredQuantity:       1.0000,
			EnteredUOMDisplayCode: "Dosage Form",
			AmountReferenced:      10.0000,
			AmountReferencedUnits: "mg",
			BegInventory:          16.0000,
			EndInventory:          15.0000,
			WitnessName:           "None/Unknown",
			WitnessID:             "Unknown",
			MRN:                   "1234567",
		},
		{
			ItemTransactionKey:    uuid.New(),
			UserName:              "Testnurse, Three",
			UserID:                "abc3c",
			StorageSpace:          "TESTPYXIS_MAIN Drw 2.1-Pkt A2",
			ItemID:                "05018",
			MedClassCode:          "Non-Controlled",
			MedDisplayName:        "diphenhydrAMINE 50 mg (1 mL) VIAL",
			TransactionType:       "Remove",
			TxDateTime:            thirdEventTime,
			EnteredQuantity:       1.0000,
			EnteredUOMDisplayCode: "Dosage Form",
			AmountReferenced:      50.0000,
			AmountReferencedUnits: "mg",
			BegInventory:          4.0000,
			EndInventory:          3.0000,
			WitnessName:           "None/Unknown",
			WitnessID:             "Unknown",
			MRN:                   "7654321",
		},
	}))

	if len(p.PyxisEventLogs[0].Log) != 3 {
		t.Fatalf("error. AddEvents method failed to add pyxis events")
	}

	newEvents := []PyxisEvent{
		{
			ItemTransactionKey:    uuid.New(),
			UserName:              "Testnurse, One",
			UserID:                "abc1a",
			StorageSpace:          "TESTPYXIS_MAIN Drw 4.1-Pkt E5",
			ItemID:                "49375",
			MedClassCode:          "3",
			MedDisplayName:        "lorazepam 0.5 mg TABLET UD",
			TransactionType:       "Remove",
			TxDateTime:            firstEventTime.Add(24 * time.Hour),
			EnteredQuantity:       1.0000,
			EnteredUOMDisplayCode: "Dosage Form",
			AmountReferenced:      0.500,
			AmountReferencedUnits: "mg",
			BegInventory:          13.0000,
			EndInventory:          12.0000,
			WitnessName:           "None/Unknown",
			WitnessID:             "Unknown",
			MRN:                   "2143657",
		},
		{
			ItemTransactionKey:    uuid.New(),
			UserName:              "Testnurse, Two",
			UserID:                "abc2b",
			StorageSpace:          "TESTPYXIS_MAIN Drw 5.1-Pkt E6",
			ItemID:                "98367",
			MedClassCode:          "Non-Controlled",
			MedDisplayName:        "acetaminophen 325 mg TABLET UD",
			TransactionType:       "Remove",
			TxDateTime:            secondEventTime.Add(24 * time.Hour),
			EnteredQuantity:       2.0000,
			EnteredUOMDisplayCode: "Dosage Form",
			AmountReferenced:      0.500,
			AmountReferencedUnits: "mg",
			BegInventory:          97.0000,
			EndInventory:          95.0000,
			WitnessName:           "None/Unknown",
			WitnessID:             "Unknown",
			MRN:                   "9876543",
		},
	}

	p.logger.Log(p.PyxisEventLogs[0].AddPyxisEvents(newEvents))

	expectedOrder := []string{
		"oxyCODONE 5 mg TABLET UD",
		"lorazepam 0.5 mg TABLET UD",
		"PROCHLORPERAZINE 10 mg TABLET UD",
		"acetaminophen 325 mg TABLET UD",
		"diphenhydrAMINE 50 mg (1 mL) VIAL",
	}

	for i, event := range p.PyxisEventLogs[0].Log {
		if event.MedDisplayName != expectedOrder[i] {
			t.Errorf("error. expected %s. found %s", expectedOrder[i], event.MedDisplayName)
		}
	}

	err = p.saveAndUnloadPyxisEventLogs()
	if err != nil {
		t.Errorf("error saving test Pyxis event logs: %s", err.Error())
	}

	p.PyxisEventLogs = []*PyxisEventLog{}
	err = p.loadPyxisEventLogs()
	if err != nil {
		t.Errorf("error loading test Pyxis event logs: %s", err.Error())
	}

	for _, pyxisEventLog := range p.PyxisEventLogs {
		if pyxisEventLog.PyxisName == "testPyxis" {
			if len(pyxisEventLog.Log) != 5 {
				t.Errorf("error. all test Pyxis event log data not retrieved from save data. Expected: %d. Actual: %d", 5, len(p.PyxisEventLogs[0].Log))
			}

		}
	}
}

func TestParseDate(t *testing.T) {
	dateString := "01/01/2026"

	date, err := parseDate(dateString)
	if err != nil {
		t.Errorf("error. couldn't parse date: %s", err.Error())
	}

	newDateString := date.Format("2006-01-02")
	if newDateString != "2026-01-01" {
		t.Errorf("error. date not parsed correctly. expected 2026-01-01. actual %s", newDateString)
	}
}

func TestCache(t *testing.T) {
	p := initProcess()

	err := p.db.Ping()
	if err != nil {
		p.dbConnection = false
	} else {
		p.dbConnection = true
	}
	if p.dbConnection {
		p.db.Close()
	}

	startTime, err := time.Parse("2006-01-02 15:04", "2026-01-01 12:00")
	if err != nil {
		t.Errorf("failed to generate startTime: %s", err.Error())
	}

	err = p.createNewPyxisEventLog("TESTPYXIS", startTime)

	testEventTime, err := time.Parse("2006-01-02 15:04", "2026-01-01 12:01")
	if err != nil {
		t.Errorf("failed to generate testEventTime: %s", err.Error())
	}

	testEvent := database.PyxisEventResponse{
		ItemTransactionKey: uuid.New(),
		UserName: sql.NullString{
			String: "Testnurse, One",
			Valid:  true,
		},
		UserID: sql.NullString{
			String: "tst1a",
			Valid:  true,
		},
		StorageSpace: sql.NullString{
			String: "SomePyxisPocket",
			Valid:  true,
		},
		ItemID: sql.NullString{
			String: "1234567",
			Valid:  true,
		},
		MedClassCode: sql.NullString{
			String: "1234567",
			Valid:  true,
		},
		MedDisplayName: sql.NullString{
			String: "Test Medication 0 mg",
			Valid:  true,
		},
		TransactionType: sql.NullString{
			String: "Remove",
			Valid:  true,
		},
		TxDateTime: sql.NullTime{
			Time:  testEventTime,
			Valid: true,
		},
		EnteredQuantity: sql.NullFloat64{
			Float64: 1.0000,
			Valid:   true,
		},
		EnteredUOMDisplayCode: sql.NullString{
			String: "test tab",
			Valid:  true,
		},
		AmountReferenced: sql.NullFloat64{
			Float64: 0.0000,
			Valid:   true,
		},
		AmountReferencedUnits: sql.NullString{
			String: "mg",
			Valid:  true,
		},
		BegInventory: sql.NullFloat64{
			Float64: 3.0000,
			Valid:   true,
		},
		EndInventory: sql.NullFloat64{
			Float64: 2.0000,
			Valid:   true,
		},
		WitnessName: sql.NullString{
			String: "None",
			Valid:  true,
		},
		WitnessID: sql.NullString{
			String: "None",
			Valid:  true,
		},
	}

	testEvents := []database.PyxisEventResponse{
		testEvent,
	}

	data, err := json.Marshal(&testEvents)
	if err != nil {
		t.Errorf("unable to marshal testEvents: %s", err.Error())
	}

	key := "GetPyxisEventsForDeviceByDateRange"
	key += "TESTPYXIS"
	key += "2026-01-01 00:00"
	key += "2026-01-02 00:00"

	p.cache.Add(key, data)

	queryStartTime, err := time.Parse("2006-01-02 1504", "2026-01-01 0000")
	if err != nil {
		t.Errorf("error parsing queryStartTime: %s", err.Error())
	}

	queryEndTime, err := time.Parse("2006-01-02 1504", "2026-01-02 0000")
	if err != nil {
		t.Errorf("error parsing queryEndTime: %s", err.Error())
	}

	params := database.GetPyxisEventsForDeviceByDateRangeParams{
		Device: "TESTPYXIS",
		Start:  queryStartTime,
		End:    queryEndTime,
	}

	_, err = getPyxisEvents(p, params)
	if err != nil {
		t.Errorf("cache failed. getPyxisEvents not checking cache: %s", err.Error())
	}

	close(p.cacheStop)
}

func TestFunctions(t *testing.T) {
	helper := func(f func(string) (time.Time, error), dateString string) time.Time {
		dateTime, err := f(dateString)
		if err != nil {
			t.Errorf("error parsing time %s", err.Error())
		}
		return dateTime
	}

	tests := []struct {
		dateOne        time.Time
		dateTwo        time.Time
		expectedResult bool
	}{
		{
			time.Now(),
			time.Now(),
			true,
		},
		{
			helper(parseDate, "1/2/2026"),
			helper(parseDate, "1/2/2026"),
			true,
		},
		{
			helper(parseDate, "1/1/2026"),
			helper(parseDate, "1/2/2026"),
			false,
		},
	}

	for _, test := range tests {
		result := isSameDay(test.dateOne, test.dateTwo)
		if result != test.expectedResult {
			t.Errorf("fail. isSameDay result incorrect. Expected: %t. Actual: %t", test.expectedResult, result)
		}
	}

}

func TestLogger(t *testing.T) {
	logger := initProcessLogger("./test/log.txt")

	logger.LogInfo("first test. regular log info")

	innerTestString := "Inner function return test."
	outterTestString := "Outter function return test."

	logger.Log(func(s string) *logResponder {
		lr := logResponder{}
		lr.AddInfo(s)
		return &lr
	}(innerTestString))

	lr := func(s string) logResponder {
		lr := logResponder{}
		lr.AddInfo(s)
		return lr
	}(outterTestString)

	logger.Log(&lr)

	logger.Close()
}

func TestParseDose(t *testing.T) {
	tests := []struct {
		dispQty    float64
		unitName   string
		strength   string
		parsedDose float64
	}{
		{
			1.0000,
			"tablet",
			"0.5 MG",
			0.5,
		},
		{
			0.2500,
			"mL",
			"4 MG/ML",
			1.0,
		},
		{
			2.0000,
			"mL",
			"1 MG/ML",
			2.0,
		},
		{
			0.2500,
			"mL",
			"2 MG/ML",
			0.5,
		},
	}

	for _, test := range tests {
		result, err := parseDose(test.dispQty, test.unitName, test.strength)
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		if result != test.parsedDose {
			t.Errorf("Fail. Expected Result: %f. Actual: %f", test.parsedDose, result)
		}
	}
}
