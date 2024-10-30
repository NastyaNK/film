package models

type Messages struct {
	Id     int    `json:"id"`
	IdRoom int    `json:"id_room"`
	IdUser int    `json:"id_user"`
	Text   string `json:"text"`
}
