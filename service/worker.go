package service

import (
	"math/rand/v2"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type syncWorker struct {
	Name     string
	Handle   func() error
	Interval func() time.Duration
}

func (w *syncWorker) run() error {
	defer func() {
		if err := recover(); err != nil {
			hlog.Errorf("sync %v panic, err: %v", w.Name, err)
		}
	}()
	return w.Handle()
}

func (w *syncWorker) Start() {
	go func() {
		hlog.Infof("start worker: %s", w.Name)
		for {
			if err := w.run(); err != nil {
				hlog.Error("sync %v failed, err: %v", w.Name, err)
			}
			time.Sleep(w.Interval())
		}
	}()
}

var (
	syncWorkers = []*syncWorker{
		{
			Name:     "saveChatRooms",
			Handle:   saveChatRooms,
			Interval: func() time.Duration { return time.Second * time.Duration(rand.IntN(30)+29) },
		},
		{
			Name:     "checkLive",
			Handle:   checkLive,
			Interval: func() time.Duration { return time.Duration(rand.IntN(2900)+500) * time.Millisecond },
		},
		{
			Name:     "refreshNewMessages",
			Handle:   refreshNewMessages,
			Interval: func() time.Duration { return time.Duration(rand.IntN(3900)+3000) * time.Millisecond },
		},
	}
)

func Init() {
	initEmailSender()
	if err := saveChatRooms(); err != nil {
		hlog.Errorf("saveChatRooms failed, err: %v", err)
		panic(err)
	}
	hlog.Infof("saveChatRooms done, then refresh new")
	if err := refreshNewMessages(); err != nil {
		hlog.Errorf("refreshNewMessages failed, err: %v", err)
		panic(err)
	}
	startWorkers()
}

func startWorkers() {
	hlog.Info("StartWorkers")
	for _, worker := range syncWorkers {
		worker.Start()
	}
	startFfmpegWatcher()
}
