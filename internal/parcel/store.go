package parcel

import (
	"database/sql"
	"errors"
	"fmt"
)

const (
	driverName     = "sqlite"     // Название драйвера базы данных
	dataSourceName = "tracker.db" // Имя источника данных (файл базы данных)
)

// Store представляет хранилище посылок, использующее SQLite.
type Store struct {
	*sql.DB // Встраивание стандартного подключения к базе данных
}

// NewStore инициализирует новое подключение к базе данных и возвращает Store.
func NewStore() (*Store, error) {
	db, err := sql.Open(driverName, dataSourceName)

	if err != nil {
		return nil, err
	}

	return &Store{
		db,
	}, nil
}

// Add добавляет новую посылку в базу данных.
// Возвращает сгенерированный ID новой посылки или ошибку.
func (s *Store) Add(p *Parcel) (int, error) {
	r, err := s.Exec(
		"INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :addr, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("addr", p.Address),
		sql.Named("created_at", p.CreatedAt),
	)

	if err != nil {
		return 0, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetById возвращает посылку по её идентификатору.
// Возвращает ошибку, если посылка не найдена или произошла другая ошибка при чтении.
func (s *Store) GetById(id int) (*Parcel, error) {
	p := Parcel{}
	err := s.QueryRow(
		"SELECT number, client, status, address, created_at FROM parcel WHERE number = :number",
		sql.Named("number", id),
	).Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("parcel with id %d not found", id)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get parcel by id %d: %w", id, err)
	}

	return &p, nil
}

// GetByClientId возвращает список всех посылок, принадлежащих указанному клиенту.
// Возвращает ошибку, если произошли ошибки при чтении.
func (s *Store) GetByClientId(clId int) ([]Parcel, error) {
	rows, err := s.Query(
		"SELECT number, client, status, address, created_at FROM parcel WHERE client = :client",
		sql.Named("client", clId),
	)

	if err != nil {
		return nil, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	var res []Parcel
	for rows.Next() {
		p := Parcel{}
		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		res = append(res, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// SetStatus обновляет статус посылки с заданным идентификатором.
// Возвращает ошибку, если статус недопустим или обновление не удалось.
func (s *Store) SetStatus(id int, status string) error {
	st := Status(status)
	if !st.IsValid() {
		return fmt.Errorf("invalid status: %s", st)
	}

	res, err := s.Exec(
		"UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", st),
		sql.Named("number", id),
	)

	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not determine affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("parcel with id %d not found", id)
	}

	return nil
}

// SetAddress обновляет адрес доставки посылки.
// Адрес можно изменить только если посылка находится в статусе "registered".
func (s *Store) SetAddress(id int, addr string) error {
	st, err := s.statusById(id)

	if err != nil {
		return err
	}

	if st != Registered {
		return fmt.Errorf("cannot change address — parcel not in registered state (current: %s)", st)
	}

	_, err = s.Exec(
		"UPDATE parcel SET address = :address WHERE number = :number",
		sql.Named("address", addr),
		sql.Named("number", id),
	)

	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	return nil
}

// DeleteById удаляет посылку из базы данных, если она находится в статусе "registered".
// Если посылка находится в другом статусе — она не будет удалена.
func (s *Store) DeleteById(id int) error {
	st, err := s.statusById(id)

	if err != nil {
		return err
	}

	if st != Registered {
		// fmt.Printf("cannot delete — parcel not in registered state (current: %s)\n", st)
		return nil
	}

	_, err = s.Exec(
		"DELETE FROM parcel WHERE number = :number",
		sql.Named("number", id),
	)

	if err != nil {
		return fmt.Errorf("failed to delete parcel: %w", err)
	}

	return nil
}

// statusById возвращает текущий статус посылки по её ID.
// Используется для проверки перед обновлением или удалением.
func (s *Store) statusById(id int) (Status, error) {
	var st Status
	err := s.QueryRow(
		"SELECT status FROM parcel WHERE number = :number",
		sql.Named("number", id),
	).Scan(&st)

	if errors.Is(err, sql.ErrNoRows) {
		return Empty, fmt.Errorf("parcel with id %d not found", id)
	}

	if err != nil {
		return Empty, fmt.Errorf("failed to get parcel status: %w", err)
	}

	return st, nil
}
