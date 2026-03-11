package app

type Building struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type Jobs struct {
	Id          int    `db:"id"`
	Description string `db:"description"`
	BuildingId  int    `db:"building_id"`
}
