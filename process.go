package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/PharmacyDoc2018/control_event_parser/cache"
	"github.com/PharmacyDoc2018/control_event_parser/cli"
	"github.com/PharmacyDoc2018/control_event_parser/database"
	"github.com/joho/godotenv"
)

const cacheInterval = 60 * time.Minute
const minPyxisEventRecheckInterval = 24 * time.Hour
const pyxisEventLogsFolder = "pyxis_event_logs"
const pyxisEventLogSettingsFolder = "log_settings"
const controlEventLogsFolder = "control_logs"

type Process struct {
	PyxisEvents          []*PyxisEvent
	MarActions           []*MarAction
	selectedEventActions *SelectedEventActions
	FacilityPyxisList    *PyxisList
	pathToData           string
	pathToSettings       string
	pathToOut            string
	logger               processLogger
	testMarActionRes     []database.MarActionResponse
	state                *processState
	settings             *Settings
	erxs                 *ERxDict
	itemIDs              *ItemIdDict
	erxItemIdLinks       *ERxItemIdLinks
	departmentCoverage   *DepartmentCoverage
	db                   *sql.DB
	dbq                  *database.Queries
	cliConfig            *cli.Config
	cache                *cache.Cache
	cacheStop            chan struct{}
	dbConnection         bool
}

func initProcess() *Process {
	p := Process{}
	p.state = initProcessState()

	connString := ""
	processLogPath := ""

	err := godotenv.Load(".env")
	if err != nil {
		err := p.initialLaunchSetup()
		if err != nil {
			fmt.Printf("ERROR. first time initilization process not complete: %s\n", err.Error())
		}
	}

	connString = os.Getenv("CONNSTRING")
	processLogPath = os.Getenv("PROCESSLOGPATH")
	p.pathToData = os.Getenv("DATAPATH")
	p.pathToSettings = os.Getenv("SETTINGSPATH")
	p.pathToOut = os.Getenv("OUTPATH")

	processLogger := initProcessLogger(processLogPath)
	p.logger = processLogger
	p.logger.LogInfo("Application Started")

	err = p.loadSettings()
	if err != nil {
		fmt.Println(err.Error())
	}
	p.logger.printToIO = p.settings.PrintLogsToCliIO

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		fmt.Printf("error creating connection pool: %s\n ", err.Error())
	}

	p.db = db
	p.dbq = database.New(db)

	p.cacheStop = make(chan struct{})
	p.cache = cache.NewCache(cacheInterval, p.cacheStop)

	fmt.Println("Attempting to connect to database...")
	err = p.db.Ping()
	if err != nil {
		fmt.Println("connection failed!")
		fmt.Println("warning: no database connection")
		p.state.DbConnectionFail()
		p.logger.LogError("Connection to database not successful")
	} else {
		fmt.Println("connection successful")
		p.state.DbConnectionSuccessful()
		p.logger.LogInfo("Connection to database successful")
	}

	p.FacilityPyxisList = initPyxisList()
	err, lr := p.FacilityPyxisList.Load(&p)
	p.logger.Log(lr)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		p.state.PyxisListSuccessful()
	}

	p.erxItemIdLinks = initERxItemIdLink()
	err = p.loadERxItemIdLinks()
	if err != nil {
		fmt.Println(err.Error())
	}

	p.departmentCoverage = initDepartmentCoverage()
	err = p.loadDepartmentCoverage()
	if err != nil {
		fmt.Println(err.Error())
	}

	p.erxs = initERXs()
	err = p.erxs.Load(&p)
	if err != nil {
		fmt.Println(err.Error())
	}

	p.itemIDs = initItemIDs()
	err, lr = p.itemIDs.Load(p.pathToData)
	p.logger.Log(lr)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		p.state.ItemIDsLoadSuccessful()
	}

	p.selectedEventActions = &SelectedEventActions{
		Map: map[EventTrailItem]struct{}{},
	}

	return &p
}

func (p *Process) save() {
	p.logger.LogInfo("Saving data...")

	if p.state.PyxisListOkay() {
		err, rl := p.FacilityPyxisList.Save(p.pathToData)
		p.logger.Log(rl)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	if p.state.ERxItemIdLinksOkay() {
		err := p.saveERxItemIdLinks()
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error while saving: %s", err.Error()))
			fmt.Println(err.Error())
		}
	} else {
		p.logger.LogInfo("ERx - ItemId links not being saved due to previous load error")
	}

	if p.state.DepartmentCoverageOkay() {
		err := p.saveDepartmentCoverage()
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error while saving: %s", err.Error()))
			fmt.Println(err.Error())
		}
	} else {
		p.logger.LogInfo("Department - Coverage links not being saved due to previous load error")
	}

	if p.state.ERXsOkay() {
		err := p.erxs.Save(p)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error while saving: %s", err.Error()))
			fmt.Println(err.Error())
		}
	} else {
		p.logger.LogInfo("ERXs not being saved due to previous load error")
	}

	if p.state.ItemIDsOkay() {
		err, lr := p.itemIDs.Save(p.pathToData)
		p.logger.Log(lr)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		p.logger.LogInfo("ItemIDs not being saved due to previous load error")
	}

}

func (p *Process) exit() {
	p.save()

	p.logger.LogInfo("Closing Application...")

	p.cliConfig.Rl.Close()
	close(p.cacheStop)
	time.Sleep(500 * time.Millisecond)
	p.logger.LogInfo("Application Closed")
	p.logger.EndSpace()
	p.logger.Close()
	os.Exit(0)
}

const defaultEnv = `CONNSTRING=""
PROCESSLOGPATH="./logs/process_log.txt"
DATAPATH="./data/"
SETTINGSPATH="./settings/"
OUTPATH="./out/"
`

func (p *Process) initialLaunchSetup() error {
	env, err := os.OpenFile(".env", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, err = env.WriteString(defaultEnv)
	if err != nil {
		return err
	}

	env.Close()
	godotenv.Load(".env")
	p.pathToData = os.Getenv("DATAPATH")
	p.pathToSettings = os.Getenv("SETTINGSPATH")
	p.pathToOut = os.Getenv("OUTPATH")

	err = os.Mkdir("./logs/", 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir("./data/", 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir(p.pathToSettings, 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir(p.pathToOut, 0755)
	if err != nil {
		return err
	}

	f, err := os.OpenFile("./data/ERxItemIdLinks.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = f.WriteString("{}")
	if err != nil {
		return err
	}
	f.Close()

	f, err = os.OpenFile(filepath.Join(p.pathToData, departmentCoverageFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = f.WriteString("{}")
	if err != nil {
		return err
	}
	f.Close()

	f, err = os.OpenFile(filepath.Join(p.pathToData, PyxisListFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = f.WriteString("{}")
	if err != nil {
		return err
	}
	f.Close()

	err = os.Mkdir("./data/log_settings/", 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir(filepath.Join(p.pathToData, controlEventLogsFolder), 0755)
	if err != nil {
		return err
	}

	return nil

}
