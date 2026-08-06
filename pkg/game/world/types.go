package world

type RealmStatus string

const (
	RealmStatusLocked   RealmStatus = "LOCKED"
	RealmStatusActive   RealmStatus = "ACTIVE"
	RealmStatusComplete RealmStatus = "COMPLETE"
)
