package main

import (
	"github.com/joho/godotenv"
	"vk-rest/configs"
	"vk-rest/configs/logger"
	"vk-rest/service/delivery/http"
	"vk-rest/service/usecase/core"
	"vk-rest/service/usecase/worker"
)

func main() {
	log := logger.GetLogger()
	err := godotenv.Load()
	if err != nil {
		log.Errorf("load .env error: %s", err.Error())
		return
	}

	psxCfg, err := configs.GetPsxConfig()
	if err != nil {
		log.Error("Create profile config error: ", err)
		return
	}

	redisCfg, err := configs.GetRedisConfig()
	if err != nil {
		log.Error("Create redis config error: ", err)
		return
	}

	core, err := usecase.GetCore(psxCfg, redisCfg, log)
	if err != nil {
		log.Error("Create core error: ", err)
		return
	}

	api := delivery.GetApi(core, log)

	w, err := worker.GetWorker(psxCfg, log)
	if err != nil {
		log.Error("Create worker error: ", err.Error())
		return
	}

	err = w.StartWorker()
	if err != nil {
		log.Error("Start worker error: ", err.Error())
		return
	}

	log.Info("Server running")
	err = api.ListenAndServe("8081")
	if err != nil {
		log.Error("ListenAndServe error: ", err)
		return
	}

}
