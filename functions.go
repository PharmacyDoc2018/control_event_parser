package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

func parseDate(dateString string) (time.Time, error) {
	dateFormats := []string{
		"2006-01-02 15:04",
		"01/02/2006 15:04",
		"1/2/2006 15:04",
		"01/02/06 15:04",
		"1/2/06 15:04",
		"01-02-2006 15:04",
		"1-2-2006 15:04",
		"01-02-06 15:04",
		"1-2-06 15:04",
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"01/02/06",
		"1/2/06",
		"01-02-2006",
		"1-2-2006",
		"01-02-06",
		"1-2-06",
	}

	for _, format := range dateFormats {
		date, err := time.Parse(format, dateString)
		if err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("error. %s not a valid date format", dateString)
}

func parseMonth(monthString string) (int, error) {
	monthInt := 0

	switch monthString {
	case "January", "Jan", "january", "jan", "1":
		monthInt = 1

	case "February", "Feb", "february", "feb", "2":
		monthInt = 2

	case "March", "Mar", "march", "mar", "3":
		monthInt = 3

	case "April", "Apr", "april", "apr", "4":
		monthInt = 4

	case "May", "may", "5":
		monthInt = 5

	case "June", "june", "Jun", "jun", "6":
		monthInt = 6

	case "July", "july", "Jul", "jul", "7":
		monthInt = 7

	case "August", "august", "Aug", "aug":
		monthInt = 8

	case "September", "september", "Sept", "sept":
		monthInt = 9

	case "October", "october", "Oct", "oct":
		monthInt = 10

	case "November", "november", "Nov", "nov":
		monthInt = 11

	case "December", "december", "Dec", "dec":
		monthInt = 12

	default:
		return monthInt, fmt.Errorf("error. unable to parse %s as a month", monthString)
	}

	return monthInt, nil
}

func timeToday() time.Time {
	now := time.Now()
	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		now.Location(),
	)

	return today
}

func timeStartDay(t time.Time) time.Time {
	startDay := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		0,
		0,
		0,
		0,
		t.Location(),
	)

	return startDay
}

func timeEndDay(t time.Time) time.Time {
	startDay := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		23,
		59,
		59,
		0,
		t.Location(),
	)

	return startDay
}

func isSameDay(timeOne, timeTwo time.Time) bool {
	return timeOne.Year() == timeTwo.Year() &&
		timeOne.Month() == timeTwo.Month() &&
		timeOne.Day() == timeTwo.Day()
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func addFloat(x, y float64) float64 {
	m := 10000.0

	xmInt := int(x * m)
	ymInt := int(y * m)

	zmInt := xmInt + ymInt
	return float64(zmInt) / m

}

func subtractFloat(x, y float64) float64 {
	m := 10000.0

	xmInt := int(x * m)
	ymInt := int(y * m)

	zmInt := xmInt - ymInt
	return float64(zmInt) / m

}

func quickDisplayName(f func(string) (string, *logError), s string) string {
	r, e := f(s)
	if e != nil {
		return "unknown"
	}
	return r
}

func printfln(format string, a ...any) (n int, err error) {
	b := a
	n, err = fmt.Println(fmt.Sprintf(format, b...))
	return n, err
}

func sortEventTrailItems(items []EventTrailItem) {
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
			dateTimeTwo = items[j].PyxisEvent.TxDateTime

		case marAction:
			dateTimeTwo = items[j].MarAction.SavedTime

		}

		return dateTimeOne.Before(dateTimeTwo)
	})
}

func printEventTrailItems(items []EventTrailItem) {
	for _, item := range items {
		switch item.Type {
		case pyxisEvent:
			printfln("Item Transaction Key: %s", item.PyxisEvent.ItemTransactionKey.String())
			printfln("Transaction Type: %s", item.PyxisEvent.TransactionType)
			printfln("Date Time: %s", item.PyxisEvent.TxDateTime.Format("2006-01-02 1504"))
			printfln("User ID: %s", item.PyxisEvent.UserID)
			printfln("User Name: %s", item.PyxisEvent.UserName)
			printfln("Display Name: %s", item.PyxisEvent.MedDisplayName)
			printfln("Amount: %s %s",
				strconv.FormatFloat(item.PyxisEvent.AmountReferenced, 'f', -1, 64),
				item.PyxisEvent.AmountReferencedUnits)
			printfln("MAR: %s", item.PyxisEvent.MRN)
			printfln("Witness: %s", item.PyxisEvent.WitnessName)
			fmt.Println()

		case marAction:
			printfln("Order Number: %s", item.MarAction.OrderNumber)
			printfln("MAR Action: %s", item.MarAction.MarAction)
			printfln("Saved Time: %s", item.MarAction.SavedTime.Format("2006-01-02 1504"))
			printfln("User ID: %s", item.MarAction.UserID)
			printfln("User Name: %s", item.MarAction.UserName)
			printfln("Display Name: %s", item.MarAction.DisplayName)
			printfln("Dose: %s %s",
				strconv.FormatFloat(item.MarAction.Dose, 'f', -1, 64),
				item.MarAction.DoseUnitDescription)
			printfln("MAR: %s", item.MarAction.MRN)
			printfln("Patient Name: %s", item.MarAction.PtName)
			fmt.Println()

		}
	}
}

func openInDefaultEditor(filePath string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows: Launches the registered default app for .txt
		cmd = exec.Command("cmd", "/c", "start", filePath)
	case "darwin":
		// macOS: Opens using the default utility
		cmd = exec.Command("open", filePath)
	case "linux":
		// Linux: Opens using the desktop environment's default app
		cmd = exec.Command("xdg-open", filePath)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start() // Use .Start() so your Go app doesn't freeze while the GUI is open
}

func parseDose(dispQty float64, unitName, strength string) (float64, error) {
	doseRunes := map[rune]struct{}{
		'0': struct{}{},
		'1': struct{}{},
		'2': struct{}{},
		'3': struct{}{},
		'4': struct{}{},
		'5': struct{}{},
		'6': struct{}{},
		'7': struct{}{},
		'8': struct{}{},
		'9': struct{}{},
		'.': struct{}{},
		',': struct{}{},
	}

	strFloat := 0.0
	var err error

	splitStrength := strings.Split(strength, "/")
	switch len(splitStrength) {
	case 0:
		return 0.0, fmt.Errorf("error. invalid strength format")

	//-- Not a concentration
	case 1:
		//-- Get the number from the start of the strength string
		strRunes := []rune(strength)
		i := 0
		_, okay := doseRunes[strRunes[i]]
		for okay {
			i++
			_, okay = doseRunes[strRunes[i]]
		}
		strRunes = strRunes[:i]

		//-- Filter out commas
		tempRunes := strRunes
		strRunes = []rune{}
		for i := range tempRunes {
			if tempRunes[i] != ',' {
				strRunes = append(strRunes, tempRunes[i])
			}
		}

		strFloat, err = strconv.ParseFloat(string(strRunes), 64)
		if err != nil {
			return 0.0, err
		}

	case 2:
		//
		strRunes := []rune(splitStrength[0])
		denominator := strings.TrimSpace(strings.ToLower(splitStrength[1]))
		if denominator != strings.TrimSpace(strings.ToLower(unitName)) {
			return 0.0, fmt.Errorf("error. %s in concentration %s does not match disp_qty unit %s", denominator, strength, unitName)
		}

		i := 0
		_, okay := doseRunes[strRunes[i]]
		for okay {
			i++
			_, okay = doseRunes[strRunes[i]]
		}
		strRunes = strRunes[:i]

		//-- Filter out commas
		tempRunes := strRunes
		strRunes = []rune{}
		for i := range tempRunes {
			if tempRunes[i] != ',' {
				strRunes = append(strRunes, tempRunes[i])
			}
		}

		strFloat, err = strconv.ParseFloat(string(strRunes), 64)
		if err != nil {
			return 0.0, err
		}

	default:
		return 0.0, fmt.Errorf("error. invalid strength format")
	}

	return multiplyFloat(dispQty, strFloat), nil
}

func multiplyFloat(x, y float64) float64 {
	m := 10000.0

	xmInt := int(x * m)
	ymInt := int(y * m)

	zmInt := xmInt * ymInt
	return float64(zmInt) / (m * m)

}

func printMemStats() {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)

	fmt.Printf("Alloc = %.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Printf("TotalAlloc = %.2f MB\n", float64(m.TotalAlloc)/1024/1024)
	fmt.Printf("HeapAlloc= %.2f MB\n", float64(m.HeapAlloc)/1024/1024)
	fmt.Printf("HeapInuse= %.2f MB\n", float64(m.HeapInuse)/1024/1024)
	fmt.Printf("HeapIdle= %.2f MB\n", float64(m.HeapIdle)/1024/1024)
	fmt.Printf("HeapReleased= %.2f MB\n", float64(m.HeapReleased)/1024/1024)
	fmt.Printf("Sys = %.2f MB\n", float64(m.Sys)/1024/1024)
	fmt.Printf("Num Gc = %d\n", m.NumGC)
}
