package parcel

import (
	"fmt"
)

// Status представляет текущий статус посылки.
type Status string

// Возможные значения статуса посылки.
const (
	Registered Status = "registered" // Посылка зарегистрирована
	Sent       Status = "sent"       // Посылка отправлена
	Delivered  Status = "delivered"  // Посылка доставлена
	Empty      Status = ""           // Пустой или неизвестный статус
)

// parcelStatusTransitions определяет допустимые переходы между статусами посылки.
var parcelStatusTransitions = map[Status]Status{
	Registered: Sent,
	Sent:       Delivered,
	Delivered:  Empty,
}

// IsValid проверяет, является ли статус допустимым (имеет следующий переход).
func (s Status) IsValid() bool {
	_, exists := parcelStatusTransitions[s]
	return exists
}

// Next возвращает следующий статус для текущего.
// Если статус невалидный — возвращается ошибка.
// Если статус равен Empty, то возвращается также Empty и ошибка не возникает.
func (s Status) Next() (Status, error) {
	if s == Empty {
		return Empty, nil
	}

	if !s.IsValid() {
		return Empty, fmt.Errorf("unknown parcel status: %s", s)
	}

	return parcelStatusTransitions[s], nil
}
