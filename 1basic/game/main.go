package main

import (
	"fmt"
	"strings"
)

/*
	код писать в этом файле
	наверняка у вас будут какие-то структуры с методами, глобальные перменные ( тут можно ), функции
*/

type Room struct {
	ID   int
	Name string

	Items           map[string]bool
	AccessibleRooms []string //куда можно пройти
	State           string
	Description     string
	StatementCheck  func()
}

type Person struct {
	CurrentRoomID int
	Inventory     []string
}
type World struct {
	Person              Person
	Rooms               []Room
	InterectableObjects []InterectableObject
}
type InterectableObject struct {
	ObjectName string
	AppliedObj string
	State      string
	RoomID     int
}

var MyWorld = World{
	Person{},
	[]Room{},
	[]InterectableObject{},
}

func (p *Person) WalkTo(room string) {

	p.CurrentRoomID = GetRoomIDByName(room)
}

func (p *Person) TakeItemToInventory(item string) {
	p.Inventory = append(p.Inventory, item)
}

func GetCommand(w *World, command []string) string {
	switch command[0] {
	case "осмотреться":
		return w.Rooms[w.Person.CurrentRoomID].State + "можно пройти - " + SliceUnfolding(w.Rooms[w.Person.CurrentRoomID].AccessibleRooms)

	case "идти":
		for _, room := range w.Rooms[w.Person.CurrentRoomID].AccessibleRooms {

			if command[1] == room {
				if command[1] == "улица" && w.InterectableObjects[0].State == "дверь закрыта" && w.Person.CurrentRoomID == 1 {
					return w.InterectableObjects[0].State
				}
				w.Person.WalkTo(room)
				return w.Rooms[w.Person.CurrentRoomID].Description + "можно пройти - " + SliceUnfolding(w.Rooms[w.Person.CurrentRoomID].AccessibleRooms)
			}
		}
		return "нет пути в " + command[1]
	case "надеть":
		for k, _ := range w.Rooms[w.Person.CurrentRoomID].Items {

			if k == command[1] {
				delete(w.Rooms[w.Person.CurrentRoomID].Items, command[1])
				w.Person.TakeItemToInventory(command[1])
				for _, v := range w.Rooms {
					v.StatementCheck()
				}
				return "вы надели: " + command[1]
			}
		}

		return "нет такого"
	case "взять":
		for k, _ := range w.Rooms[w.Person.CurrentRoomID].Items {

			if k == command[1] {
				if IsValueInList("рюкзак", w.Person.Inventory) {
					delete(w.Rooms[w.Person.CurrentRoomID].Items, command[1])
					w.Person.TakeItemToInventory(command[1])
					for _, v := range w.Rooms {
						v.StatementCheck()
					}

					return "предмет добавлен в инвентарь: " + command[1]
				} else {
					return "некуда класть"
				}
			}
		}

		return "нет такого"

	case "применить":
		if IsValueInList(command[1], w.Person.Inventory) {
			for _, v := range w.InterectableObjects {
				if v.ObjectName == command[1] && v.AppliedObj == command[2] {
					if v.ObjectName == "ключи" {
						w.InterectableObjects[0].State = "дверь открыта"

						return "дверь открыта"
					}
				} else {
					return "не к чему применить"
				}
			}
		} else {
			return "нет предмета в инвентаре - " + command[1]
		}

	}

	return "неизвестная команда"
}
func GetRoomIDByName(name string) (ID int) {
	for _, v := range MyWorld.Rooms {
		if v.Name == name {
			return v.ID
		}
	}
	return 0
}

func SliceUnfolding(sl []string) string {
	UnfoldedSlice := ""
	for i := 0; i < len(sl)-1; i++ {
		UnfoldedSlice = UnfoldedSlice + sl[i] + ", "
	}
	UnfoldedSlice = UnfoldedSlice + sl[len(sl)-1]
	return UnfoldedSlice
}
func IsValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func main() {
	/*
		в этой функции можно ничего не писать
		но тогда у вас не будет работать через go run main.go
		очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/

	initGame()
	fmt.Println(handleCommand("идти коридор"))
	fmt.Println(handleCommand("идти комната"))
	fmt.Println(handleCommand("надеть рюкзак"))
	fmt.Println(handleCommand("взять ключи"))
	fmt.Println(handleCommand("идти коридор"))
	fmt.Println(handleCommand("применить ключи дверь"))
	fmt.Println(handleCommand("идти улица"))
}
func initGame() {
	/*
		эта функция инициализирует игровой мир - все команты
		если что-то было - оно корректно перезатирается
	*/

	//затирание
	MyWorld = World{
		Person{},
		[]Room{},
		[]InterectableObject{},
	}

	MyWorld.Rooms = append(MyWorld.Rooms,
		Room{ID: 0, Name: "кухня", Items: nil, AccessibleRooms: []string{"коридор"}, State: "ты находишься на кухне, на столе: чай, надо собрать рюкзак и идти в универ. ", Description: "кухня, ничего интересного. ", StatementCheck: func() {}},
		Room{ID: 1, Name: "коридор", Items: nil, AccessibleRooms: []string{"кухня", "комната", "улица"}, Description: "ничего интересного. ", StatementCheck: func() {}},
		Room{ID: 2, Name: "улица", Items: nil, AccessibleRooms: []string{"домой"}, Description: "на улице весна. ", StatementCheck: func() {}},
		Room{ID: 3, Name: "комната", Items: map[string]bool{
			"рюкзак":    true,
			"ключи":     true,
			"конспекты": true,
		}, AccessibleRooms: []string{"коридор"}, State: "на столе: ключи, конспекты, на стуле: рюкзак. ", Description: "ты в своей комнате. ", StatementCheck: func() {}})

	MyWorld.Person.CurrentRoomID = 0
	MyWorld.Person.Inventory = nil
	MyWorld.InterectableObjects = append(MyWorld.InterectableObjects, InterectableObject{"ключи", "дверь", "дверь закрыта", 1})

	MyWorld.Rooms[0].StatementCheck = func() {
		w := &MyWorld.Rooms[0]
		p := &MyWorld.Person
		if IsValueInList("рюкзак", p.Inventory) {
			w.State = "ты находишься на кухне, на столе: чай, надо идти в универ. "
		} else {
			w.State = "ты находишься на кухне, на столе: чай, надо собрать рюкзак и идти в универ. "
		}

	}
	MyWorld.Rooms[3].StatementCheck = func() {
		room := &MyWorld.Rooms[3]
		if len(room.Items) == 0 {
			room.State = "пустая комната. "
		} else {
			room.State = "на столе: "
			for i, _ := range room.Items {
				if i == "ключи" || i == "конспекты" {
					room.State = room.State + i + ", "
				}
				if i == "рюкзак" {
					room.State = room.State + "на стуле: " + i
				}
			}

		}
		room.State = room.State[:len(room.State)-2] + ". "

	}

}
func handleCommand(command string) string {
	/*
		данная функция принимает команду от "пользователя"
		и наверняка вызывает какой-то другой метод или функцию у "мира" - списка комнат
	*/

	CommandToPerson := strings.Split(command, " ")
	answer := GetCommand(&MyWorld, CommandToPerson)

	return answer
}
