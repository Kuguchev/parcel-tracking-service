package parcel

import (
	"database/sql"
	"errors"
	"fmt"
)

const (
	driverName     = "sqlite"     // Название драйвера базы данных
	DataSourceName = "tracker.db" // Имя источника данных (файл базы данных)
)

// Repository представляет хранилище посылок, использующее SQLite.
type Repository struct {
	*sql.DB // Встраивание стандартного подключения к базе данных
}

// NewRepository инициализирует новое подключение к базе данных и возвращает Repository.
func NewRepository(dataSourceName string) (*Repository, error) {
	db, err := sql.Open(driverName, dataSourceName)

	if err != nil {
		return nil, err
	}

	return &Repository{
		db,
	}, nil
}

// Add добавляет новую посылку в базу данных.
// Возвращает сгенерированный ID новой посылки или ошибку.
func (r *Repository) Add(p *Parcel) (int, error) {
	res, err := r.Exec(
		"INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :addr, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("addr", p.Address),
		sql.Named("created_at", p.CreatedAt),
	)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Get возвращает посылку по её идентификатору.
// Возвращает ошибку, если посылка не найдена или произошла другая ошибка при чтении.
func (r *Repository) Get(id int) (*Parcel, error) {
	p := Parcel{}
	err := r.QueryRow(
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
func (r *Repository) GetByClientId(id int) ([]*Parcel, error) {
	rows, err := r.Query(
		"SELECT number, client, status, address, created_at FROM parcel WHERE client = :client",
		sql.Named("client", id),
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

	var res []*Parcel
	for rows.Next() {
		p := Parcel{}
		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		res = append(res, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// SetStatus обновляет статус посылки с заданным идентификатором.
// Возвращает ошибку, если статус недопустим или обновление не удалось.
func (r *Repository) SetStatus(id int, status string) error {
	st := Status(status)
	if !st.IsValid() {
		return fmt.Errorf("invalid status: %s", st)
	}

	res, err := r.Exec(
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
func (r *Repository) SetAddress(id int, addr string) error {
	res, err := r.Exec(
		"UPDATE parcel SET address = :address WHERE number = :number AND status = :status",
		sql.Named("address", addr),
		sql.Named("number", id),
		sql.Named("status", Registered),
	)

	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not determine affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("parcel with id %d not found or not in registered status", id)
	}

	return nil
}

// Delete удаляет посылку из базы данных по её идентификатору и статусу.
//
// Параметры:
//   - id: идентификатор посылки (поле `number` в БД).
//   - status: ожидаемый статус посылки для удаления. Если передан `Empty` — статус не проверяется.
//
// Возвращает:
//   - ошибку, если удаление не удалось (например, из-за несоответствия статуса или ошибки выполнения запроса).
func (r *Repository) Delete(id int, status Status) error {
	query := "DELETE FROM parcel WHERE number = :number"
	args := []any{sql.Named("number", id)}

	if status != Empty {
		query += " AND status = :status"
		args = append(args, sql.Named("status", status))
	}

	_, err := r.Exec(query, args...)

	if err != nil {
		return fmt.Errorf("failed to delete parcel: %w", err)
	}

	return nil
}
