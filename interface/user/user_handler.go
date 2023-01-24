package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"forum/app"
	"forum/domain/entity"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

type UserInfo struct {
	userApp app.UserAppInterface
	logger  *zap.Logger
}

func NewUserInfo(userApp app.UserAppInterface,
	logger *zap.Logger) *UserInfo {
	return &UserInfo{
		userApp: userApp,
		logger:  logger,
	}
}

func (userInfo *UserInfo) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	userInfo.logger.Info("HandleCreateUser")
	vars := mux.Vars(r)
	nickname := vars[string(entity.NicknameKey)]

	user := &entity.User{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		userInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(data, user)
	if err != nil {
		userInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user.Nickname = nickname

	err = userInfo.userApp.CreateUser(user)
	if err != nil {
		users, err := userInfo.userApp.GetUsersWithNicknameAndEmail(nickname, user.Email)
		if err != nil {
			userInfo.logger.Info(
				err.Error(), zap.String("url", r.RequestURI),
				zap.String("method", r.Method))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err := json.Marshal(users)
		if err != nil {
			userInfo.logger.Info(
				err.Error(), zap.String("url", r.RequestURI),
				zap.String("method", r.Method))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(body)
		return
	}

	body, err := json.Marshal(user)
	if err != nil {
		userInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
	return
}

func (userInfo *UserInfo) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	userInfo.logger.Info("HandleGetUser")
	vars := mux.Vars(r)
	nickname := vars[string(entity.NicknameKey)]

	profile, err := userInfo.userApp.GetUserByNickname(nickname)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find user with id #%v\n", nickname),
		}
		body, err := json.Marshal(msg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(body)
		return
	}

	body, err := json.Marshal(profile)
	if err != nil {
		userInfo.logger.Info(
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

func (userInfo *UserInfo) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userInfo.logger.Info("HandleUpdateUser")
	vars := mux.Vars(r)
	nickname := vars[string(entity.NicknameKey)]

	profile := &entity.User{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		userInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(data, profile)
	if err != nil {
		userInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	profile.Nickname = nickname

	profileData, err := userInfo.userApp.UpdateUser(profile)
	if err != nil {
		var msg entity.Message
		if errors.Is(err, entity.UserDoesntExistsError) {
			msg = entity.Message{
				Text: fmt.Sprintf("Can't find user with id #%v\n", nickname),
			}
			body, err := json.Marshal(msg)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write(body)
			return
		} else if errors.Is(err, entity.DataError) {
			emailOwnerNickname, err := userInfo.userApp.GetUserNicknameWithEmail(profile.Email)
			if err != nil {
				userInfo.logger.Info(
					err.Error(), zap.String("url", r.RequestURI),
					zap.String("method", r.Method))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			msg = entity.Message{
				Text: fmt.Sprintf("This email is already registered by user: %v", emailOwnerNickname),
			}
			body, err := json.Marshal(msg)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write(body)
			return
		}
	}

	body, err := json.Marshal(profileData)
	if err != nil {
		userInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
