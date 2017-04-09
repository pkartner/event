package event

type TimelineStore struct {
	stores []Store
	currentTime uint64
	lastEventID ID
}

type NewBranchEvent struct {
	PrevBranch ID
	PrevBranchLastEvent ID
}

func NewBranch(prevBranch ID, prevBranchLastEvent ID) *Event {
	return &Event{
		Attributes: &NewBranchEvent{
			PrevBranch: prevBranch, 
			PrevBranchLastEvent: prevBranchLastEvent,
		},
		Type: "new_branch",
        ID: ID([16]byte("something")),
        Time: 0,
        StoreID: "timeline",
	}
}

type WindbackEvent struct {
	Time uint64
}

func Windback(time uint64) *Event {
	return &Event{
		Attributes: &WindbackEvent{
			Time: time,
		},
		Type: "new_branch",
        ID: ID([16]byte("something")),
        Time: 0,
        StoreID: "timeline",
	}
}

type Branch struct {
    CreationTime uint64 
    BranchID ID 
    PreviousBranch *Branch
    PreviousBranchLastEventId ID
    StoreID ID
}

// func (b *Branch) NewBranch(time int, previousBranchLastEventId ID, store* Store) *Branch {
//     return &Branch{
//         CreationTime: time, 
//         BranchID: 0,
//         PreviousBranch: b,
//         PreviousBranchLastEventId: previousBranchLastEventId,
//         StoreID: store.ID,
//     }
// }

// type Timeline struct {
//     MainBranchID ID
    
// }