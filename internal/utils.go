package internal

import "log"

// LogError imprime un error si ocurre
func LogError(err error) {
	if err != nil {
		log.Println("Error:", err)
	}
}
