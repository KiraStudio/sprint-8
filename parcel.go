package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	var id int
	query := `INSERT INTO parcel (client, status, address, created_at) 
              VALUES ($1, $2, $3, $4) 
              RETURNING number`
	err := s.db.QueryRow(query, p.Client, p.Status, p.Address, p.CreatedAt).Scan(&id)
	if err != nil {
		// fallback для windows , у меня всегда выводило 0
		res, err2 := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES ($1, $2, $3, $4)",
			p.Client, p.Status, p.Address, p.CreatedAt)
		if err2 != nil {
			return 0, err2
		}
		id64, _ := res.LastInsertId()
		if id64 != 0 {
			return int(id64), nil
		}
		err2 = s.db.QueryRow("SELECT last_insert_rowid()").Scan(&id)
		if err2 != nil {
			return 0, fmt.Errorf("insert failed and cannot retrieve id: %w", err2)
		}
	}
	return id, nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	p := Parcel{}
	row := s.db.QueryRow("SELECT client, status, address, created_at FROM parcel WHERE number = $1", number)
	err := row.Scan(&p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = $1", client)
	if err != nil {
		return nil, err
	}
	var res []Parcel
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET status = $1 WHERE number = $2", status, number)
	if err != nil {
		return err
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	currentStatus := ""
	row := s.db.QueryRow("SELECT status FROM parcel WHERE number = $1", number)
	err := row.Scan(&currentStatus)
	if err != nil {
		return err
	}
	if currentStatus == ParcelStatusRegistered {
		_, err := s.db.Exec("UPDATE parcel SET address = $1 WHERE number = $2", address, number)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	_, err := s.db.Exec("DELETE FROM parcel WHERE number = $1 and status = $2", number, ParcelStatusRegistered)
	if err != nil {
		return err
	}
	return nil
}
