package services

func Run() {
	InitConfig()

	RegisterHealthCheckServer()

	RegisterGoRoutines()
}
