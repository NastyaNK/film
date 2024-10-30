package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	. "m/models"
	http "net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var jwtKey = []byte("ключ") //ключ, который используется для подписи и проверки

type Service struct {
	repo *Repository
}

//	func searchFilms(db *sqlx.DB, filmNames string) error {
//		query := "SELECT name, year, rating,image FROM film WHERE name LIKE '%' || $1 || '%'"
//		rows, err := db.Query(query, filmNames)
//		if err != nil {
//			return err
//		}
//		var film Film
//		for rows.Next() {
//			err := rows.Scan(&film.Name, &film.Year, &film.rating, &film.image)
//			if err != nil {
//				return err
//			}
//			fmt.Printf("Movie: %s, Year: %d, Rating: %.1f, image: %s\n", film.Name, film.Year, film.rating, film.image)
//		}
//		rows.Close()
//		return nil
//	}
//
//	func addFilm(db *sqlx.DB, film Film) error {
//		_, err := db.Exec("insert into film (name,year,plot,genre,rating,image)values ($1,$2,$3,$4,$5,$6)", film.Name, film.Year, film.Plot, film.Genre, film.rating, film.image)
//		if err != nil {
//			return err
//		}
//		return err
//	}
//
//	func deleteFilm(db *sqlx.DB, film Film) error {
//		_, err := db.Exec("delete from film WHERE (name=$1)", film.Name)
//		if err != nil {
//			return err
//		}
//		return err
//	}
func (service *Service) registration(w http.ResponseWriter, r *http.Request) { //(service *Service) привязывание функции к структуре

	if r.Header.Get("Content-Type") != "application/json" { //проверяем что в заголовке данные именно json
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	var people People
	bytes, err := io.ReadAll(r.Body)     //получаем все данные из заголовка (адресной строки)
	err = json.Unmarshal(bytes, &people) //распихиваем по структуре
	if err != nil {
		http.Error(w, "Ошибка в чтении"+err.Error(), http.StatusBadRequest)
		return
	} //в случае ошибки показать ее статус
	re := regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+={}[\]:;'",.<->/?]+$`) //присваимваем регулярное выражение
	if !re.MatchString(people.Password) {                                   //проверяем что пароль не содержит ничего кроме выражения
		http.Error(w, "Пароль содержит недопустимые символы используйте только буквы (a-z, A-Z), цифры (0-9) и специальные символы (!@#$%^&*()).", http.StatusBadRequest)
	}
	re = regexp.MustCompile(`^.{8,20}$`)  //функция MustCompile получает регулярное выражение(где ^. начало строки $ конец строки)
	if !re.MatchString(people.Password) { //функция MatchString  для проверки соответствия регулярного выражения
		http.Error(w, "Длина пароля должна быть от 8 до 20 символов", http.StatusBadRequest)
		return
	}
	exist, err := service.repo.existUser(people)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exist {
		http.Error(w, "Пользователь с таким именем или почтой уже существует", http.StatusBadRequest)
		return
	}
	err = service.repo.addUser(people)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func createToken(people People) (string, error) {
	// Создаем новый токен с указанием метода подписи (SigningMethodHS256) и с тем что будет содержаться в нем (MapClaims - мапа)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": people,                             // В мапу добавляем пользователя
		"exp":  time.Now().AddDate(0, 0, 1).Unix(), // В мапу добавляем срок дейстивя токена (1 день)
	})
	return token.SignedString(jwtKey) // Подписываем токен с использованием секретного ключа
}

func validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // Если подписано неизвестным методом получай ошибку
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	//if !token.Valid { // Проверка срока действия токена
	//	return nil, errors.New("ошибка токен недействителен")
	//}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token structure")
	}
	return claims, err
}

// Перевод мапы в структуру (m - мапа из токена,p - ссылка на структуру)
func mapToStruct(m interface{}, p interface{}) error {
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, p)
	if err != nil {
		return err
	}
	return nil
}

//func v()
//	claims, err := validateToken(tokenString)
//	if err != nil {
//
//	}
//	var peple People
//	err = mapToStruct(claims["user"], &peple)
//	if err != nil {
//		return
//	}
//}

func (service *Service) Auth(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") != "application/json" { //проверяем что в заголовке данные именно json
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	var people People
	bytes, err := io.ReadAll(r.Body)     //получаем все данные из заголовка (адресной строки)
	err = json.Unmarshal(bytes, &people) //распихиваем по структуре
	if err != nil {
		http.Error(w, "Некорректный запрос "+err.Error(), http.StatusBadRequest)
		return
	} //в случае ошибки показать ее статус

	people1, err := service.repo.getUserByName(people.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(people1.Password), []byte(people.Password))
	if err != nil {
		http.Error(w, "Неверный пароль: "+err.Error(), http.StatusUnauthorized) // Если имя пользователя или пароль неверны, возвращаем ошибку аутентификации
		return
	}

	tokenString, err := createToken(people1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")  //показывает в заголовке что возвращаем в том же формате
	response := map[string]string{"token": tokenString} //создаем мапу с токеном
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Отказ в доступе "+err.Error(), http.StatusUnauthorized)
		return
	} //преобразуем в json и отправляем обратно клиенту в адресную строку

}
func (service *Service) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var room Room
	tokenString := r.Header.Get("Authorization") // Получаем токен из заголовка
	claims, err := validateToken(tokenString)
	if err != nil {
		http.Error(w, "Ошибка в проверке токена2 "+err.Error(), http.StatusUnauthorized)
		return
	}
	var user People
	err = mapToStruct(claims["user"], &user) // преобразование заявки "user" в структуру People
	if err != nil {
		http.Error(w, "Токен неверен"+err.Error(), http.StatusUnauthorized)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка в чтении "+err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bytes, &room)
	if err != nil {
		http.Error(w, "Ошибка в декодировании JSON "+err.Error(), http.StatusBadRequest)
		return
	}
	err = service.repo.addRoom(room, user)
	if err != nil {
		http.Error(w, "Неверный запрос "+err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Комната успешно создана"))
}

func (service *Service) participants(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" { //проверяем что в заголовке данные именно json
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}
	tokenString := r.Header.Get("Authorization") // Получаем токен из заголовка
	claims, err := validateToken(tokenString)
	if err != nil {
		http.Error(w, "Ошибка в проверке токена2 "+err.Error(), http.StatusUnauthorized)
		return
	}
	var people People
	err = mapToStruct(claims["user"], &people) // преобразование заявки "people" в структуру People
	if err != nil {
		http.Error(w, "Токен неверен "+err.Error(), http.StatusUnauthorized)
		return
	}
	var userRoom UserRoom
	bytes, err := io.ReadAll(r.Body)       //получаем все данные из заголовка (адресной строки)
	err = json.Unmarshal(bytes, &userRoom) //распихиваем по структуре
	if err != nil {
		http.Error(w, "Некорректный запрос "+err.Error(), http.StatusBadRequest)
		return
	}
	room, err := service.repo.getRoom(userRoom.IdRoom)
	if err != nil {
		http.Error(w, "Комната не найдена"+err.Error(), http.StatusNotFound)
		return
	}

	userRoom, err = service.repo.getParticipant(people, room)
	if room.Public {
		if err != nil {
			err = service.repo.addParticipant(UserRoom{
				IdRoom:    room.Id,
				IdUser:    people.Id,
				Role:      "PARTICIPANT",
				IsInvited: false,
				Ban:       false,
			})
			if err != nil {
				http.Error(w, "Внутренняя ошибка "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else if userRoom.Ban && !userRoom.IsInvited {
			http.Error(w, "Вы в бане", http.StatusMethodNotAllowed)
			return
		} else {
			userRoom.IsInvited = false
			userRoom.Ban = false
			err := service.repo.updateParticipant(userRoom)
			if err != nil {
				http.Error(w, "Внутренняя ошибка "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		if err != nil {
			http.Error(w, "Комната закрыта", http.StatusMethodNotAllowed)
			return
		} else if userRoom.IsInvited {
			userRoom.IsInvited = false
			userRoom.Ban = false
			err := service.repo.updateParticipant(userRoom)
			if err != nil {
				http.Error(w, "Внутренняя ошибка "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else if userRoom.Ban {
			http.Error(w, "Вы в бане", http.StatusMethodNotAllowed)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}
func (service *Service) invitation(w http.ResponseWriter, r *http.Request) {
	// Проверка заголовка на тип данных JSON
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Проверка авторизации пользователя
	tokenString := r.Header.Get("Authorization")
	claims, err := validateToken(tokenString)
	if err != nil {
		http.Error(w, "Ошибка в проверке токена "+err.Error(), http.StatusUnauthorized)
		return
	}
	var people People
	err = mapToStruct(claims["user"], &people)
	if err != nil {
		http.Error(w, "Токена нет "+err.Error(), http.StatusUnauthorized)
		return
	}
	var userRoom UserRoom
	err = json.NewDecoder(r.Body).Decode(&userRoom) //чтение параметра из тела UserRoom
	if err != nil {
		http.Error(w, "Некорректные данные "+err.Error(), http.StatusBadRequest)
		return
	}

	// Получение данных о комнате
	room, err := service.repo.getRoom(userRoom.IdRoom)
	if err != nil {
		http.Error(w, "Комната не найдена "+err.Error(), http.StatusNotFound)
		return
	}
	user, err := service.repo.getUser(userRoom.IdUser)
	if err != nil {
		http.Error(w, "Человека нет "+err.Error(), http.StatusNotFound)
		return
	}
	participant, err := service.repo.getParticipant(people, room)
	if err != nil {
		http.Error(w, "любая ошибка "+err.Error(), http.StatusNotFound)
	}
	if participant.Role != "ADMIN" {
		http.Error(w, "Вы не можете приглашать участников", http.StatusBadRequest)
		return
	}
	invitation, err := service.repo.getParticipant(user, room)
	if err != nil {
		err = service.repo.addParticipant(UserRoom{
			IdRoom:    room.Id,
			IdUser:    user.Id,
			Role:      "PARTICIPANT",
			IsInvited: true,
			Ban:       false,
		})
		if err != nil {
			http.Error(w, "добавление не удалось "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		if invitation.Ban {
			invitation.Ban = false
			invitation.IsInvited = true
			err = service.repo.updateParticipant(invitation)
			if err != nil {
				http.Error(w, "обновление не удалось "+err.Error(), http.StatusBadRequest)
				return
			}
		}
	}
	// Отправка успешного ответа
	w.WriteHeader(http.StatusOK)

}
func (service *Service) ban(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	tokenString := r.Header.Get("Authorization")
	claims, err := validateToken(tokenString)
	if err != nil {
		http.Error(w, "Ошибка в проверке токена "+err.Error(), http.StatusUnauthorized)
		return
	}

	var people People
	err = mapToStruct(claims["user"], &people)
	if err != nil {
		http.Error(w, "Токена нет "+err.Error(), http.StatusUnauthorized)
		return
	}

	var userRoom UserRoom
	err = json.NewDecoder(r.Body).Decode(&userRoom)
	if err != nil {
		http.Error(w, "Некорректные данные "+err.Error(), http.StatusBadRequest)
		return
	}
	if people.Id == userRoom.IdUser {
		http.Error(w, "Вы не можете себя забанить", http.StatusBadRequest)
		return
	}
	room, err := service.repo.getRoom(userRoom.IdRoom)
	if err != nil {
		http.Error(w, "Комната не найдена "+err.Error(), http.StatusNotFound)
		return
	}

	participant, err := service.repo.getParticipant(people, room)
	if err != nil {
		http.Error(w, "Человека нет "+err.Error(), http.StatusNotFound)
		return
	}

	if participant.Role != "ADMIN" {
		http.Error(w, "Вы не можете добавлять в бан", http.StatusBadRequest)
		return
	}

	banParticipant, err := service.repo.getParticipant(People{Id: userRoom.IdUser}, room)
	if err != nil {
		http.Error(w, "Ошибка при получении информации о пользователе для бана "+err.Error(), http.StatusInternalServerError)
		return
	}

	if banParticipant.Ban {
		http.Error(w, "Пользователь уже находится в бане", http.StatusForbidden)
		return
	}

	banParticipant.Ban = true
	err = service.repo.updateParticipant(banParticipant)
	if err != nil {
		http.Error(w, "Ошибка при обновлении статуса бана "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func (service *Service) exitRom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Неподходящий метод ", http.StatusBadRequest)
		return
	}
	tokenString := r.Header.Get("Authorization")
	claims, err := validateToken(tokenString)
	if err != nil {
		http.Error(w, "Ошибка в проверке токена "+err.Error(), http.StatusUnauthorized)
		return
	}
	var people People
	err = mapToStruct(claims["user"], &people)
	if err != nil {
		http.Error(w, "Токена нет "+err.Error(), http.StatusUnauthorized)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	var userRoom UserRoom
	err = json.NewDecoder(r.Body).Decode(&userRoom)
	if err != nil {
		http.Error(w, "Некорректные данные "+err.Error(), http.StatusBadRequest)
		return
	}
	user, err := service.repo.getParticipant(people, Room{Id: userRoom.IdRoom})
	if err != nil {
		http.Error(w, "Не удалось достать пользователя "+err.Error(), http.StatusNotFound)
		return
	}
	err = service.repo.deleteUserRoom(user.IdUser, userRoom.IdRoom)
	if err != nil {
		http.Error(w, "Не удалось удалить пользователя "+err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	host := "localhost"
	port := "9893"

	db := connect()

	service := Service{repo: db} //создаем в структуре переменную с подключением к базе

	router := chi.NewRouter() //Роутер в веб-приложении - это компонент, который обрабатывает HTTP-запросы и перенаправляет их на
	// соответствующие обработчики в зависимости от URL-адреса запроса и метода (GET, POST и т.д.).

	// http://localhost:9893/*

	router.Post("/auth", service.Auth) // Определение обработчика POST-запросов на "/auth"
	router.Post("/registration", service.registration)
	router.Post("/createRoom", service.CreateRoom)
	router.Post("/participant", service.participants)
	router.Post("/invitation", service.invitation)
	router.Post("/ban", service.ban)
	router.Post("/exit", service.exitRom)

	root := http.Dir("./front")      //указываем где находятся все нужные мне веб страницы
	router.Get("/*", funcName(root)) //Эта строка говорит роутеру обрабатывать все GET-запросы (запросы на чтение данных),
	// которые приходят на любой путь (обозначенный здесь как "/*"),

	log.Println("Server started http://" + host + ":" + port)
	log.Fatal(http.ListenAndServe(host+":"+port, router))
}

func funcName(root http.Dir) func(w http.ResponseWriter, r *http.Request) { //принимает root (путь к папке с веб-страницами) и
	// возвращает другую функцию, которая может обрабатывать веб-запросы.
	return func(w http.ResponseWriter, r *http.Request) {
		routerCtx := chi.RouteContext(r.Context()) //Это получение контекста маршрутизации из запроса. Это помогает узнать, какой путь
		// запрашивается. “www.yoursite.com/about”. Здесь “/about” - это путь
		pathPrefix := strings.TrimSuffix(routerCtx.RoutePattern(), "/*") //Это удаление суффикса "/*" из паттерна маршрута. Это делается,
		// чтобы получить основной путь
		fs := http.StripPrefix(pathPrefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //Это создание обработчика файлов,
			// который получает основной путь “www.yoursite.com/about/team”, то после удаления “/about”, останется только “/team”, и сервер будет искать файл “team” в папке “about”.
			path := "./front" + r.URL.Path                                      //Это создание полного пути к файлу, который нужно отдать. Он состоит из папки ‘front’ и пути из URL запроса.
			if file, err := os.Stat(path); os.IsNotExist(err) || file.IsDir() { //Это проверка, существует ли файл по этому пути и не является ли он директорией.
				http.NotFound(w, r) // Если файл не найден или является директорией, то сервер отправляет ответ “404 Not Found”.
				return
			}
			http.FileServer(root).ServeHTTP(w, r) //Если файл существует и не является директорией, то сервер отдает этот файл.
		}))
		fs.ServeHTTP(w, r) //чтение файла с диска и отправку его содержимого обратно пользователю
	}
}
