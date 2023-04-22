package status_updaters

import (
	vmModel "go-deploy/models/vm"
	"go-deploy/pkg/app"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func vmStatusUpdater(ctx *app.Context) {
	for {
		if ctx.Stop {
			break
		}

		allVms, _ := vmModel.GetAll()
		for _, vm := range allVms {
			code, message, err := fetchVmStatus(&vm)
			if err != nil {
				continue
			}
			_ = vmModel.UpdateByID(vm.ID, bson.D{{"statusCode", code}, {"statusMessage", message}})
		}

		time.Sleep(1 * time.Second)
	}
}

func deploymentStatusUpdater(ctx *app.Context) {
	for {
		if ctx.Stop {
			break
		}

		time.Sleep(1 * time.Second)
	}
}
