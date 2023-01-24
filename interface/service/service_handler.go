package service

import (
	"encoding/json"
	"fmt"
	"forum/app"
	"forum/domain/entity"
	"go.uber.org/zap"
	"net/http"
)

type ServiceInfo struct {
	ServiceApp app.ServiceAppInterface
	logger     *zap.Logger
}

func NewServiceInfo(ServiceApp app.ServiceAppInterface, logger *zap.Logger) *ServiceInfo {
	return &ServiceInfo{
		ServiceApp: ServiceApp,
		logger:     logger,
	}
}

func (serviceInfo *ServiceInfo) HandleClearData(w http.ResponseWriter, r *http.Request) {
	serviceInfo.logger.Info("HandleClearData")
	err := serviceInfo.ServiceApp.ClearAllDate()
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf(`{"messege": "%s"}`, err.Error()),
		}
		body, err := json.Marshal(msg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(body)
		return
	}

	body, err := json.Marshal("")
	if err != nil {
		serviceInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (serviceInfo *ServiceInfo) HandleGetDBStatus(w http.ResponseWriter, r *http.Request) {
	serviceInfo.logger.Info("HandleGetDBStatus")

	status, err := serviceInfo.ServiceApp.GetDBStatus()
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf(`{"messege": "%s"}`, err.Error()),
		}
		body, err := json.Marshal(msg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(body)
		return
	}

	body, err := json.Marshal(status)
	if err != nil {
		serviceInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
	return
}
