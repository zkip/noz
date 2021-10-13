package utils

func Setup() func() {
	LoadConfig()

	return Clean
}

func Clean() {
	cleanSQL()
}
