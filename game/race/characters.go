package race

type Character struct {
	ID            string `json:"_id" bson:"_id"`
	Theme         string `json:"theme" bson:"theme"`
	GuildID       string `json:"guild_id" bson:"guild_id"`
	Emoji         string `json:"emoji" bson:"emoji"`
	MovementSpeed string `json:"movement_speed" bson:"movement_speed"`
}
