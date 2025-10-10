package parcel

import (
	"fmt"
	"time"
)

type Parcel struct {
	Number                     int
	Client                     int
	Status, Address, CreatedAt string
}

type Service struct {
	Store
}

func NewService(store *Store) *Service {
	return &Service{
		*store,
	}
}

func (s *Service) Register(clId int, addr string) (*Parcel, error) {
	p := Parcel{
		Client:    clId,
		Status:    string(Registered),
		Address:   addr,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	id, err := s.Add(&p)
	if err != nil {
		return nil, err
	}

	p.Number = id

	fmt.Printf("Новая посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		p.Number, p.Address, p.Client, p.CreatedAt)

	return &p, nil
}

func (s *Service) PrintClientParcels(clId int) error {
	parcels, err := s.GetByClientId(clId)

	if err != nil {
		return err
	}

	fmt.Printf("Посылки клиента %d:\n", clId)

	for _, p := range parcels {
		fmt.Printf("Посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			p.Number, p.Address, p.Client, p.CreatedAt, p.Status)
	}

	fmt.Println()

	return nil
}

func (s *Service) NextStatus(id int) error {
	p, err := s.GetById(id)
	if err != nil {
		return err
	}

	nextStatus, err := Status(p.Status).Next()
	if err != nil {
		return err
	}

	if nextStatus == Empty {
		return nil
	}

	fmt.Printf("У посылки № %d новый статус: %s\n", id, nextStatus)

	return s.SetStatus(id, string(nextStatus))
}

func (s *Service) ChangeAddr(id int, addr string) error {
	return s.SetAddress(id, addr)
}

func (s *Service) Delete(id int) error {
	return s.DeleteById(id)
}
