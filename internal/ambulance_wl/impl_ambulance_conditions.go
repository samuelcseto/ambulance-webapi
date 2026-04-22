package ambulance_wl

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (this *implAmbulanceConditionsApi) GetConditions(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}
