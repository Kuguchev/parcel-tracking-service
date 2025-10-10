package main

import (
	"fmt"

	"github.com/Kuguchev/parcel-tracking-service/internal/parcel"
	_ "modernc.org/sqlite"
)

func main() {
	s, err := parcel.NewStore()

	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		err := s.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	ps := parcel.NewService(s)

	// регистрация посылки
	clientId := 1
	addr := "Псков, д. Пушкина, ул. Колотушкина, д. 5"
	p, err := ps.Register(clientId, addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// изменение адреса
	newAddr := "Саратов, д. Верхние Зори, ул. Козлова, д. 25"
	err = ps.ChangeAddr(p.Number, newAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// изменение статуса
	err = ps.NextStatus(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// вывод посылок клиента
	err = ps.PrintClientParcels(clientId)
	if err != nil {
		fmt.Println(err)
		return
	}

	// попытка удаления отправленной посылки
	err = ps.Delete(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// вывод посылок клиента
	// предыдущая посылка не должна удалиться, т.к. её статус НЕ «зарегистрирована»
	err = ps.PrintClientParcels(clientId)
	if err != nil {
		fmt.Println(err)
		return
	}

	// регистрация новой посылки
	p, err = ps.Register(clientId, addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// удаление новой посылки
	err = ps.Delete(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// вывод посылок клиента
	// здесь не должно быть последней посылки, т.к. она должна была успешно удалиться
	err = ps.PrintClientParcels(clientId)
	if err != nil {
		fmt.Println(err)
		return
	}
}
