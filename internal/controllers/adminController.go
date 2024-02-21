package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"palyvoua/internal/models"
	"palyvoua/tools/auth"
	"palyvoua/tools/jsonHelper"
)

type adminController struct {
	adminRepo adminRepo
	userRepo userRepository
}

type adminRepo interface {
	GetAllRoles() ([]models.Role, error)
	SaveRole(role models.Role) error
	DeleteRoleByID(roleID uuid.UUID) error
	GetRoleByID(uuid.UUID) (models.Role, error)
	GetRoleByName(string) (models.Role, error)
}

func SetupAdminRoutes(r *gin.Engine, ar adminRepo, ur userRepository) {

	adminGroup := r.Group("/admin")

	ac := adminController{ar, ur}

	adminGroup.GET("/role/byId",)

	adminGroup.Use(auth.AuthMiddleware(ur))
	adminGroup.Use(auth.RoleMiddleware(3, ur ,ar))
	adminGroup.GET("/role", jsonHelper.MakeHttpHandler(ac.getAllRoles))
	adminGroup.POST("/role", jsonHelper.MakeHttpHandler(ac.createRole))
	adminGroup.DELETE("/role/:id", jsonHelper.MakeHttpHandler(ac.deleteRoleByID))
}

func (ac *adminController) getRoleByID(c *gin.Context) error {

	idField := c.Query("id")
	id, err := uuid.Parse(idField)
	if err != nil {
		return err
	}
	role, err := ac.adminRepo.GetRoleByID(id)
	if err != nil {
		return err
	}
	c.JSON(200, gin.H{"role":role})
	return nil
}

func (ac *adminController) getAllRoles(c *gin.Context) error {

	roles, err := ac.adminRepo.GetAllRoles()
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error retrieving roles",
			Status: 500,
		}
	}
	c.JSON(200, gin.H{"roles": roles})
	return nil
}

type CreateRoleRequest struct {
	Name string `bson:"name" json:"name"`
	AuthorityLevel int `json:"authorityLevel" bson:"authorityLevel"`
}

func (ac *adminController) createRole(c *gin.Context) error {

	var body CreateRoleRequest
	if err:=c.Bind(&body);err!=nil {

		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}
	roleID, err := uuid.NewRandom()
	if err != nil {
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}
	role := models.Role{
		ID: roleID,
		Name:           body.Name,
		AuthorityLevel: body.AuthorityLevel,
	}

	err = ac.adminRepo.SaveRole(role)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error saving role",
			Status: 500,
		}
	}

	return nil
}

func (ac *adminController) deleteRoleByID(c *gin.Context) error {
	roleToDelete := c.Param("id")
	err := ac.adminRepo.DeleteRoleByID(uuid.MustParse(roleToDelete))
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error deleting role with given id",
			Status: 500,
		}
	}
	c.JSON(200, gin.H{})
	return nil
}
