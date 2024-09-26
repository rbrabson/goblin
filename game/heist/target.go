package heist

// Target is a target of a heist.
type Target struct {
	ID       string  `json:"_id" bson:"_id"`
	CrewSize int64   `json:"crew" bson:"crew"`
	Success  float64 `json:"success" bson:"success"`
	Vault    int64   `json:"vault" bson:"vault"`
	VaultMax int64   `json:"vault_max" bson:"vault_max"`
}
