package library

// i am only including objects and fields i care about for
// a simple bot

type Snowflake string

type User struct {
	ID              Snowflake `json:"id"`
	Username 	  string    `json:"username"`
	Discriminator  string    `json:"discriminator"`
	Avatar         string    `json:"avatar"`
	Bot            bool      `json:"bot"`
	Email          string    `json:"email"`
}

type GuildMember struct {
	User           *User      `json:"user"`
	Nick           *string    `json:"nick"`
	Roles          []Snowflake `json:"roles"`
	JoinedAt       string    `json:"joined_at"`
	Avatar		 *string    `json:"avatar"`
}

type Channel struct {
	ID              Snowflake `json:"id"`
	GuildID         *Snowflake `json:"guild_id"`
	Type            int       `json:"type"`
	Name 		  *string    `json:"name"`
}

type Guild struct {
	ID              Snowflake `json:"id"`
	Name            string    `json:"name"`
	Icon            *string    `json:"icon"`
	OwnerID         Snowflake `json:"owner_id"`
	Channels 	  *[]Channel `json:"channels"`
	Members 	  *[]GuildMember `json:"members"`
	Unavailable     bool      `json:"unavailable"`
}

type Message struct {
	ID              Snowflake `json:"id"`
	ChannelID       Snowflake `json:"channel_id"`
	GuildID         *Snowflake `json:"guild_id"`
	Author          *User     `json:"author"`
	Content         string    `json:"content"`
}
