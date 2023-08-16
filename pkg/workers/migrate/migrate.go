package migrator

import "log"

// This file is edited when an update in the database schema has occurred.
// Thus, every migration needed will be done programmatically.
// Once a migration is done, clear the file.

// Migrate will  run as early as possible in the program, and it will never be called again.
func Migrate() {
	log.Println("nothing to migrate")
}
