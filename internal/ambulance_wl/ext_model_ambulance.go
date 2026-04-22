package ambulance_wl

import (
	"slices"
	"time"
)

func (this *Ambulance) reconcileWaitingList() {
	if len(this.WaitingList) == 0 {
		return
	}

	slices.SortFunc(this.WaitingList, func(left, right WaitingListEntry) int {
		if left.WaitingSince.Before(right.WaitingSince) {
			return -1
		} else if left.WaitingSince.After(right.WaitingSince) {
			return 1
		}
		return 0
	})

	if this.WaitingList[0].EstimatedStart.Before(this.WaitingList[0].WaitingSince) {
		this.WaitingList[0].EstimatedStart = this.WaitingList[0].WaitingSince
	}

	if this.WaitingList[0].EstimatedStart.Before(time.Now()) {
		this.WaitingList[0].EstimatedStart = time.Now()
	}

	nextEntryStart := this.WaitingList[0].EstimatedStart.
		Add(time.Duration(this.WaitingList[0].EstimatedDurationMinutes) * time.Minute)

	for i := 1; i < len(this.WaitingList); i++ {
		entry := &this.WaitingList[i]
		if entry.EstimatedStart.Before(nextEntryStart) {
			entry.EstimatedStart = nextEntryStart
		}
		if entry.EstimatedStart.Before(entry.WaitingSince) {
			entry.EstimatedStart = entry.WaitingSince
		}
		nextEntryStart = entry.EstimatedStart.
			Add(time.Duration(entry.EstimatedDurationMinutes) * time.Minute)
	}
}
