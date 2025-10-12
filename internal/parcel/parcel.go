package parcel

// Parcel представляет посылку, зарегистрированную в системе доставки.
type Parcel struct {
	Number                     int
	Client                     int
	Status, Address, CreatedAt string
}
