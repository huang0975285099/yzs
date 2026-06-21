package handlers

import (
	"net/http"

	"go-yzs/database"
	"go-yzs/models"

	"github.com/gin-gonic/gin"
)

type CreateTeamRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type UpdateTeamRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

func ListTeams(c *gin.Context) {
	var teams []models.Team
	database.DB.Order("id asc").Find(&teams)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": teams})
}

func CreateTeam(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	var count int64
	database.DB.Model(&models.Team{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "团队名称已存在"})
		return
	}

	team := models.Team{Name: req.Name}
	if err := database.DB.Create(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建团队失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": team})
}

func UpdateTeam(c *gin.Context) {
	id := c.Param("id")
	var team models.Team
	if err := database.DB.First(&team, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "团队不存在"})
		return
	}

	var req UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	var count int64
	database.DB.Model(&models.Team{}).Where("name = ? AND id != ?", req.Name, team.ID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "团队名称已存在"})
		return
	}

	team.Name = req.Name
	database.DB.Save(&team)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": team})
}

func DeleteTeam(c *gin.Context) {
	id := c.Param("id")
	var team models.Team
	if err := database.DB.First(&team, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "团队不存在"})
		return
	}

	// Clear team assignment from users before deleting
	database.DB.Model(&models.User{}).Where("team_id = ?", team.ID).Update("team_id", nil)
	database.DB.Delete(&team)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
