package library

type Ready struct {
	Version           int                 `json:"v"`
	SessionID         string              `json:"session_id"`
	Shard             []int               `json:"shard"`
	User              User               `json:"user"`
	Guilds            []*Guild            `json:"guilds"`
}


