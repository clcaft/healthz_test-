package dto

type RabbitSendRequest struct {
	Message string `json:"message" binding:"required" example:"hello rabbit"`
	Source  string `json:"source" binding:"required" example:"go-service"`
}

type RabbitSendResponse struct {
	Message string `json:"message" example:"sent"`
}

type RabbitMessage struct {
	Message   string `json:"message" example:"hello rabbit"`
	Source    string `json:"source" example:"go-service"`
	CreatedAt string `json:"created_at" example:"2026-05-25T18:30:00+05:00"`
}
