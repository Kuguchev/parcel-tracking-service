package parcel

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// TODO: при параллельном запуске тестов обернуть попробовать каждый тест в транзакицию.. Ну и нужно больше тестов:)

const TestDataSourceName = "../../tracker_test.db"

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() *Parcel {
	return &Parcel{
		Client:    1000,
		Status:    string(Registered),
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	pr, err := NewRepository(TestDataSourceName)
	require.NoError(t, err)

	defer func() {
		err := pr.Close()
		require.NoError(t, err)
	}()

	tp := getTestParcel()

	id, err := pr.Add(tp)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	p, err := pr.Get(id)
	require.NoError(t, err)
	tp.Number = p.Number
	assert.Equal(t, *tp, *p)

	err = pr.Delete(id, Status(tp.Status))
	require.NoError(t, err)

	_, err = pr.Get(id)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	pr, err := NewRepository(TestDataSourceName)
	require.NoError(t, err)

	defer func() {
		err := pr.Close()
		require.NoError(t, err)
	}()

	tp := getTestParcel()

	id, err := pr.Add(tp)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	defer func() {
		_ = pr.Delete(id, Status(tp.Status))
	}()

	newAddr := "new test address"
	err = pr.SetAddress(id, newAddr)
	require.NoError(t, err)

	p, err := pr.Get(id)
	require.NoError(t, err)
	tp.Number = p.Number
	tp.Address = newAddr
	assert.Equal(t, *tp, *p)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	pr, err := NewRepository(TestDataSourceName)
	require.NoError(t, err)

	defer func() {
		err := pr.Close()
		require.NoError(t, err)
	}()

	tp := getTestParcel()

	id, err := pr.Add(tp)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	defer func() {
		_ = pr.Delete(id, Empty)
	}()

	sentStatus := string(Sent)
	err = pr.SetStatus(id, sentStatus)
	require.NoError(t, err)

	storedParcel, err := pr.Get(id)
	require.NoError(t, err)
	assert.Equal(t, sentStatus, storedParcel.Status)

	deliveredStatus := string(Delivered)
	err = pr.SetStatus(id, deliveredStatus)
	require.NoError(t, err)

	storedParcel, err = pr.Get(id)
	require.NoError(t, err)
	assert.Equal(t, deliveredStatus, storedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	pr, err := NewRepository(TestDataSourceName)
	require.NoError(t, err)

	defer func() {
		err := pr.Close()
		require.NoError(t, err)
	}()

	parcelsNum := 3
	parcels, parcelMap := make([]*Parcel, 0, parcelsNum), make(map[int]*Parcel, parcelsNum)

	clientId := randRange.Intn(10_000_000)

	for i := 0; i < parcelsNum; i++ {
		parcels = append(parcels, getTestParcel())
		parcels[i].Client = clientId

		id, err := pr.Add(parcels[i])
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		parcels[i].Number = id

		parcelMap[id] = parcels[i]
	}

	defer func() {
		for id := range parcelMap {
			_ = pr.Delete(id, Empty)
		}
	}()

	storedParcels, err := pr.GetByClientId(clientId)
	require.NoError(t, err)
	assert.Equal(t, parcelsNum, len(storedParcels))

	for _, parcel := range storedParcels {
		testParcel, exists := parcelMap[parcel.Number]
		require.True(t, exists)
		testParcel.Number = parcel.Number
		assert.Equal(t, *testParcel, *parcel)
	}
}
