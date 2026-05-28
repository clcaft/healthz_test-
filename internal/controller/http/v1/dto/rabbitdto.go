package dto

type RabbitSendRequest struct {
	Message string `json:"message" binding:"required" example:"hello rabbit"`
	Source  string `json:"source" binding:"required" example:"go-service"`
}

type RabbitSendResponse struct {
	Message string `json:"message" example:"sent"`
}
