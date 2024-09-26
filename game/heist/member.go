package heist

import "time"

type CriminalLevel int

const (
	Greenhorn CriminalLevel = 0
	Renegade  CriminalLevel = 1
	Veteran   CriminalLevel = 10
	Commander CriminalLevel = 25
	WarChief  CriminalLevel = 50
	Legend    CriminalLevel = 75
	Immortal  CriminalLevel = 100
)

type MemberStatus int

const (
	FREE MemberStatus = iota
	DEAD
	APPREHENDED
	OOB
)

// Member is the status of a member who has participated in at least one heist
type Member struct {
	ID            string        `json:"_id" bson:"_id"`
	BailCost      int           `json:"bail_cost" bson:"bail_cost"`
	CriminalLevel CriminalLevel `json:"criminal_level" bson:"criminal_level"`
	DeathTimer    time.Time     `json:"death_timer" bson:"death_timer"`
	Deaths        int           `json:"deaths" bson:"deaths"`
	JailCounter   int           `json:"jail_counter" bson:"jail_counter"`
	Sentence      time.Duration `json:"sentence" bson:"sentence"`
	Spree         int           `json:"spree" bson:"spree"`
	Status        MemberStatus  `json:"status" bson:"status"`
	JailTimer     time.Time     `json:"time_served" bson:"time_served"`
	TotalJail     int           `json:"total_jail" bson:"total_jail"`
}

// String returns the string representation for a criminal level
func (level CriminalLevel) String() string {
	switch {
	case level >= Immortal:
		return "Immortal"
	case level >= Legend:
		return "Legend"
	case level >= WarChief:
		return "WarChief"
	case level >= Commander:
		return "Commander"
	case level >= Veteran:
		return "Veteran"
	case level >= Renegade:
		return "Renegade"
	case level >= Greenhorn:
		return "Greenhorn"
	default:
		return "Unknown"
	}
}

// String returns a string representation of the status of the member of a heist
func (status MemberStatus) String() string {
	switch status {
	case FREE:
		return "Free"
	case DEAD:
		return "Dead"
	case APPREHENDED:
		return "Apprehende"
	case OOB:
		return "Out on Bail"
	default:
		return "Unknownn"
	}
}
