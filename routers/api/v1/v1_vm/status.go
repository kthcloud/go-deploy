package v1_vm

//func GetStatus(c *gin.Context) {
//	context := app.NewContext(c)
//
//	rules := validator.MapData{
//		"vmId": []string{"required", "uuid_v4"},
//	}
//
//	validationErrors := context.ValidateParams(&rules)
//
//	if len(validationErrors) > 0 {
//		context.ResponseValidationError(validationErrors)
//		return
//	}
//
//	token, err := context.GetKeycloakToken()
//	if err != nil {
//		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
//	}
//	deploymentID := context.GinContext.Param("vmId")
//	userId := token.Sub
//
//	statusCode, deploymentStatus, _ := vm_service.GetStatusByID(userId, deploymentID)
//	if deploymentStatus == nil {
//		context.NotFound()
//		return
//	}
//
//	if statusCode == status_codes.DeploymentNotFound {
//		context.JSONResponse(http.StatusNotFound, deploymentStatus)
//		return
//	}
//
//	context.JSONResponse(http.StatusOK, deploymentStatus)
//
//}
