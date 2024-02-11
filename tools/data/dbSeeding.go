package data

type DBSeeder interface {
	SeedDB() error
}

type defaultDBSeeder struct {
}

func (s *defaultDBSeeder) SeedDB() error {
	err := s.createDefaultRoles()
	if err != nil {
		return err
	}

	return nil
}

func NewDBSeeder() DBSeeder {
	seeder := defaultDBSeeder{}
	return &seeder
}