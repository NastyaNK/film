package models

type Room struct {
	Id     int    `json:"id"`
	IdFilm int    `json:"id_film" db:"id_film"`
	Name   string `json:"name"`
	Public bool   `json:"is_public" db:"is_public"`
}
