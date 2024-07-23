package services

func Run() {
	InitConfig()

	InitRabbitMQ()

	RegisterHealthCheckServer()

	RegisterGoRoutines()
}
