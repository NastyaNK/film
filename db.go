package main

import (
	"errors"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	. "m/models"
)

var DB *sqlx.DB

func (repo *Repository) existUser(people People) (bool, error) {
	var count int
	err := repo.db.QueryRow("SELECT * FROM people WHERE name=$1 OR mail=$2", people.Name, people.Mail).Scan(&count)
	if err != nil {
		return false, errors.New("ошибка при выполнении запроса: " + err.Error())
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (repo *Repository) addUser(people People) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(people.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("ошибка при хэшировании пароля: " + err.Error())
	}
	_, err = repo.db.Exec("INSERT INTO people(name, password, mail) VALUES ($1, $2, $3)", people.Name, hashedPassword, people.Mail)
	if err != nil {
		return errors.New("ошибка при выполнении запроса: " + err.Error())
	}
	return nil
}
func (repo *Repository) getUserByName(name string) (People, error) {
	var people People
	err := repo.db.Get(&people, "SELECT * FROM people WHERE name = $1", name)
	if err != nil {
		return people, errors.New("пользователя с таким именем нет " + err.Error())
	}
	return people, nil
}

func (repo *Repository) addParticipant(userRoom UserRoom) error {
	_, err := repo.db.Exec("INSERT INTO room_user (room_id, user_id, role, is_invited, ban) VALUES ($1, $2, $3, $4, $5)", userRoom.IdRoom, userRoom.IdUser, userRoom.Role, userRoom.IsInvited, userRoom.Ban)
	if err != nil {
		return errors.New("Ошибка добавления участника: " + err.Error())
	}
	return nil
}
func (repo *Repository) getParticipant(people People, room Room) (UserRoom, error) {
	var userRoom UserRoom
	err := repo.db.Get(&userRoom, "SELECT * FROM room_user WHERE room_id = $1 AND user_id = $2", room.Id, people.Id)
	if err != nil {
		return userRoom, errors.New("Не достали участника" + err.Error())
	}
	return userRoom, nil
}
func (repo *Repository) updateParticipant(userRoom UserRoom) error {
	_, err := repo.db.Exec("update room_user SET role=$3, is_invited=$4, ban=$5 WHERE room_id=$1 and user_id=$2", userRoom.IdRoom, userRoom.IdUser, userRoom.Role, userRoom.IsInvited, userRoom.Ban)
	if err != nil {
		return errors.New("Ошибка добавления участника: " + err.Error())
	}
	return nil
}

func (repo *Repository) getUser(id int) (People, error) {
	var user People
	err := repo.db.Get(&user, "select * from people where id=$1", id)
	if err != nil {
		return user, errors.New("Ошибка получения пользователя" + err.Error())
	}
	return user, nil
}

func (repo *Repository) deleteUserRoom(idUser, idRoom int) error {
	_, err := repo.db.Exec("DELETE FROM room_user WHERE user_id=$1 and room_id=$2", idUser, idRoom)
	if err != nil {
		return errors.New("Удаление зрителя не удалось " + err.Error())
	}
	return nil
}
func (repo *Repository) deletePeople(id int) error {
	_, err := repo.db.Exec("DELETE FROM people WHERE id=$1", id)
	if err != nil {
		return errors.New("Удаление пользователя не удалось " + err.Error())
	}
	return nil
}
