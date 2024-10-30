package models

type UserRoom struct {
	Id        int    `json:"id" db:"id"`
	IdRoom    int    `json:"room_id" db:"room_id"`
	IdUser    int    `json:"user_id" db:"user_id"`
	Role      string `json:"role" db:"role"`
	IsInvited bool   `json:"is_invited" db:"is_invited"`
	Ban       bool   `json:"ban" db:"ban"`
}
