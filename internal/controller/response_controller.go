package controller

import (
	"fmt"
	"net/http"

	"aidashboard/internal/service/dashboard_service"
	"aidashboard/internal/service/openai_service"

	"github.com/gin-gonic/gin"
)

// @Summary Get AI Response
// @Description Get an AI-generated response based on the user command
// @Accept  json
// @Produce  json
// @Param  user_command body string true "User command"
// @Success 200 {string} string "Response"
// @Router /ai-dashboard [post]
func GetResponse(c *gin.Context) {
	var requestBody struct {
		UserCommand string `json:"user_command"`
	}

	// Bind the incoming JSON to the requestBody struct
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userCommand := requestBody.UserCommand
	fmt.Println(userCommand)
	if userCommand == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User command is required"})
		return
	}

	prompt := dashboard_service.GetPrompt(userCommand)
	openaiResp, err := openai_service.CallOpenAI(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dashboardResp, err := dashboard_service.GenerateDashboard(openaiResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, dashboardResp)
	}
}
