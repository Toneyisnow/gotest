package storage

import "time"

const (
	DependencyNotifierSweepTimeSpan = 10 * time.Second
)

type DependencyData struct {

	timeStamp time.Time
	data []byte
}

//
// Usage: this class is used to set data dependency for a data, when the dependent data is ready, notify to process the data
// Sample:
// notifier := NewDependencyNotifier(func(value []byte) {
//     handle_the_value(value)
// })
// notifier.SetDependency("x1", "a1")
//
// (Other thread when completed processing "x1":)
// notifier.Notify("x1")
//
type DependencyNotifier struct {

	notifyFunction func(value []byte)
	notifyMap map[string][]*DependencyData
}

func NewDependencyNotifier(notifyFunc func(value []byte)) *DependencyNotifier {

	notifier := new(DependencyNotifier)
	notifier.notifyFunction = notifyFunc
	notifier.notifyMap = make(map[string][]*DependencyData)

	return notifier
}

func (this *DependencyNotifier) SetDependency(dependent []byte, value []byte) {

	key := string(dependent)

	newData := &DependencyData{time.Now(), value}

	if existingValues, ok := this.notifyMap[key]; ok {
		this.notifyMap[key] = append(existingValues, newData)
	} else {
		this.notifyMap[key] = []*DependencyData { newData }
	}
}

func (this *DependencyNotifier) Notify(dependent []byte) {

	if this.notifyFunction == nil {
		return
	}

	key := string(dependent)
	if existingValues, ok := this.notifyMap[key]; ok {
		for _, value := range existingValues {
			if value != nil {
				this.notifyFunction(value.data)
			}
		}
	}
	delete(this.notifyMap, key)
}

func (this *DependencyNotifier) Sweep() {

	if this.notifyFunction == nil {
		return
	}

	for key, existingValues := range this.notifyMap {

		restValues := make([]*DependencyData, 0)
		for _, value := range existingValues {
			if value == nil {
				continue
			}

			if time.Now().Sub(value.timeStamp) > DependencyNotifierSweepTimeSpan {
				// For the data with old timestamp, notify it
				this.notifyFunction(value.data)
			} else {
				restValues = append(restValues, value)
			}
		}

		if len(restValues) > 0 {
			this.notifyMap[key] = restValues
		} else {
			delete(this.notifyMap, key)
		}
	}
}




