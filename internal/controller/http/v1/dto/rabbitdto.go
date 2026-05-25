package dto

type RabbitSendRequest struct {
	Queue string `json:"queue" binding:"required" example:"test-queue"`
	Data  any    `json:"data"`
}

type RabbitSendResponse struct {
	Queue string `json:"queue" example:"test-queue"`
	Data  any    `json:"data"`
}
