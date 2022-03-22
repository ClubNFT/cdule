package cdule

import (
	"os"
	"time"

	"github.com/deepaksinghvi/cdule/pkg/model"

	"gorm.io/gorm"
)

// WorkerID string
var WorkerID string

// Cdule holds watcher objects
type Cdule struct {
	*WorkerWatcher
	*ScheduleWatcher
}

func init() {
	WorkerID = getWorkerID()
}

// NewCduleWithWorker to create new scheduler with worker
func (cdule *Cdule) NewCduleWithWorker(workerName string, db *gorm.DB) error {
	WorkerID = workerName
	err := cdule.NewCdule(db)
	if err != nil {
		return err
	}
	return nil
}

// NewCdule to create new scheduler with default worker name as hostname
func (cdule *Cdule) NewCdule(db *gorm.DB) error {
	model.ConnectDataBase(db)

	worker, err := model.CduleRepos.CduleRepository.GetWorker(WorkerID)
	if nil != err {
		return err
	}
	if nil != worker {
		worker.UpdatedAt = time.Now()
		_, err := model.CduleRepos.CduleRepository.UpdateWorker(worker)
		if err != nil {
			return err
		}
	} else {
		// First time cdule started on a worker node
		worker := model.Worker{
			WorkerID:  WorkerID,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			DeletedAt: gorm.DeletedAt{},
		}
		_, err := model.CduleRepos.CduleRepository.CreateWorker(&worker)
		if err != nil {
			return err
		}
	}

	cdule.createWatcherAndWaitForSignal()
	return nil
}

func (cdule *Cdule) createWatcherAndWaitForSignal() {
	/*
		schedule watcher stop logic to abort program with signal like ctrl + c
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)*/

	workerWatcher := createWorkerWatcher()
	schedulerWatcher := createSchedulerWatcher()
	cdule.WorkerWatcher = workerWatcher
	cdule.ScheduleWatcher = schedulerWatcher
	/*select {
	case sig := <-c:
		fmt.Printf("Received %s signal. Aborting...\n", sig)
		workerWatcher.Stop()
		schedulerWatcher.Stop()
	}*/
}

// StopWatcher to stop watchers
func (cdule Cdule) StopWatcher() {
	cdule.WorkerWatcher.Stop()
	cdule.ScheduleWatcher.Stop()
}
func createWorkerWatcher() *WorkerWatcher {
	workerWatcher := &WorkerWatcher{
		Closed: make(chan struct{}),
		Ticker: time.NewTicker(time.Second * 30), // used for worker health check update in db.
	}

	workerWatcher.WG.Add(1)
	go func() {
		defer workerWatcher.WG.Done()
		workerWatcher.Run()
	}()
	return workerWatcher
}

func createSchedulerWatcher() *ScheduleWatcher {
	scheduleWatcher := &ScheduleWatcher{
		Closed: make(chan struct{}),
		Ticker: time.NewTicker(time.Minute * 1), // used for worker health check update in db.
	}

	scheduleWatcher.WG.Add(1)
	go func() {
		defer scheduleWatcher.WG.Done()
		scheduleWatcher.Run()
	}()
	return scheduleWatcher
}

func getWorkerID() string {
	hostname, err := os.Hostname()
	if err != nil {
		os.Exit(1)
	}
	return hostname
}
