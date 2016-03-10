package ImageService

import "sync"

var globalFinishedWorkMapMutex sync.Mutex

func RunService() {
	globalFinishedWorkMapMutex = sync.Mutex{}
	workToDo := make(chan string, 1024)
	finishedWorkMap := make(map[string]bool)
	go startProcessor(workToDo, &finishedWorkMap)
	setupWebInterface(workToDo, &finishedWorkMap)
}
