package repo

import (
	"errors"
	. "m"
	. "m/models"
)

func (repo *Repository) addRoom(room Room, people People) error {
	var roomID int64

	tx, err := repo.db.Beginx() //запускаем транзакцию
	if err != nil {
		return errors.New("ошибка начала транзакции: " + err.Error())
	}

	err = tx.QueryRowx("INSERT INTO room (id_film, name, is_public) VALUES ($1, $2, $3) RETURNING id", room.IdFilm, room.Name, room.Public).Scan(&roomID)
	if err != nil {
		tx.Rollback() //если возникла ошибка в запросе откатываем все изменения
		return errors.New("ошибка создания комнаты: " + err.Error())
	}
	_, err = tx.Exec("INSERT INTO room_user (room_id, user_id, role, is_invited, ban) VALUES ($1, $2, $3, $4, $5)", roomID, people.Id, "ADMIN", false, false)
	if err != nil {
		tx.Rollback()
		return errors.New("ошибка создания комнаты: " + err.Error())
	}

	return tx.Commit() //фиксирует все изменения
}

func (repo *Repository) getRoom(id int) (Room, error) {
	var room Room
	err := repo.db.Get(&room, "select * from room where id=$1", id)
	if err != nil {
		return room, errors.New("Не удалось получить комнату" + err.Error())
	}
	return room, nil
}

func (repo *Repository) deleteRoom(id int) error {
	_, err := repo.db.Exec("DELETE FROM room WHERE id=$1", id)
	if err != nil {
		return errors.New("Удаление комнаты не удалось " + err.Error())
	}
	return nil
}
