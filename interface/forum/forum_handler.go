package forum

import (
	"encoding/json"
	"fmt"
	"forum/app"
	"forum/domain/entity"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
)

type ForumInfo struct {
	ForumApp  app.ForumAppInterface
	UserApp   app.UserAppInterface
	ThreadApp app.ThreadAppInterface
	logger    *zap.Logger
}

func NewForumInfo(
	ForumApp app.ForumAppInterface,
	UserApp app.UserAppInterface,
	ThreadApp app.ThreadAppInterface,
	logger *zap.Logger) *ForumInfo {
	return &ForumInfo{
		ForumApp:  ForumApp,
		UserApp:   UserApp,
		ThreadApp: ThreadApp,
		logger:    logger,
	}
}

func (forumInfo *ForumInfo) HandleCreateForum(w http.ResponseWriter, r *http.Request) {
	forumInfo.logger.Info("HandleCreateForum")
	forum := &entity.Forum{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		forumInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(data, forum)
	if err != nil {
		forumInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nickname, err := forumInfo.UserApp.CheckIfUserExists(forum.User)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find user with id #%v\n", forum.User),
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

	forum.User = nickname

	err = forumInfo.ForumApp.CreateForum(forum)
	if err != nil {
		forumInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))

		existingForum, err := forumInfo.ForumApp.GetForumDetails(forum.Slug)
		if err != nil {
			forumInfo.logger.Info(
				err.Error(), zap.String("url", r.RequestURI),
				zap.String("method", r.Method))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err := json.Marshal(existingForum)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(body)
		return
	}

	body, err := json.Marshal(forum)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)

}

func (forumInfo *ForumInfo) HandleGetForumDetails(w http.ResponseWriter, r *http.Request) {
	forumInfo.logger.Info("HandleGetForumDetails")
	vars := mux.Vars(r)
	slug := vars[string(entity.SlugKey)]

	forum, err := forumInfo.ForumApp.GetForumDetails(slug)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find user with id #%v\n", slug),
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

	body, err := json.Marshal(forum)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (forumInfo *ForumInfo) HandleCreateForumThread(w http.ResponseWriter, r *http.Request) {
	forumInfo.logger.Info("HandleCreateForumThread")
	vars := mux.Vars(r)
	slug := vars[string(entity.SlugKey)]

	thread := &entity.Thread{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		forumInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(data, thread)
	if err != nil {
		forumInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	thread.Forum = slug

	nickname, err := forumInfo.UserApp.CheckIfUserExists(thread.Author)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find user with id #%v\n", thread.Author),
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
	thread.Author = nickname

	err = forumInfo.ThreadApp.CreateThread(thread)
	if err != nil {
		if err == entity.ForumNotExistError {
			msg := entity.Message{
				Text: fmt.Sprintf("Can't find thread forum by slug: %v", thread.Forum),
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

		existedThread, err := forumInfo.ThreadApp.GetThread(*thread.Slug)
		if err != nil {
			forumInfo.logger.Info(
				err.Error(), zap.String("url", r.RequestURI),
				zap.String("method", r.Method))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err := json.Marshal(existedThread)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(body)
		return
	}

	body, err := json.Marshal(thread)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func (forumInfo *ForumInfo) HandleGetForumUsers(w http.ResponseWriter, r *http.Request) {
	forumInfo.logger.Info("HandleGetForumUsers")
	vars := mux.Vars(r)
	slug := vars[string(entity.SlugKey)]

	_, err := forumInfo.ForumApp.CheckForumCase(slug)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find forum by slug: %v", slug),
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
	queryParams := r.URL.Query()

	limitParam, _ := queryParams[string(entity.LimitKey)]
	limit := 0
	if limitParam != nil {
		limit, err = strconv.Atoi(limitParam[0])
		if err != nil {
			forumInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	descParam, _ := queryParams[string(entity.DescKey)]
	desc := false
	if descParam == nil {
		desc = false
	} else {
		if descParam[0] == "true" {
			desc = true
		}
	}

	sinceParam, _ := queryParams[string(entity.SinceKey)]
	since := ""
	if sinceParam == nil {
		since = ""
	} else {
		since = sinceParam[0]
	}

	users, err := forumInfo.ForumApp.GetForumUsers(slug, int32(limit), since, desc)
	if err != nil {
		forumInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(users)
	if err != nil {
		forumInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (forumInfo *ForumInfo) HandleGetForumThreads(w http.ResponseWriter, r *http.Request) {
	forumInfo.logger.Info("HandleGetForumThreads")
	vars := mux.Vars(r)
	slug := vars[string(entity.SlugKey)]

	_, err := forumInfo.ForumApp.CheckForumCase(slug)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find forum by slug: %v", slug),
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

	queryParams := r.URL.Query()

	limitParam, _ := queryParams[string(entity.LimitKey)]
	limit, err := strconv.Atoi(limitParam[0])
	if err != nil {
		forumInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	descParam, _ := queryParams[string(entity.DescKey)]
	desc := false
	if descParam == nil {
		desc = false
	} else {
		if descParam[0] == "true" {
			desc = true
		}
	}

	sinceParam, _ := queryParams[string(entity.SinceKey)]
	since := ""
	if sinceParam != nil {
		since = sinceParam[0]
	}

	threads, err := forumInfo.ThreadApp.GetThreadsByForumSlug(slug, int32(limit), since, desc)
	if err != nil {
		forumInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(threads)
	if err != nil {
		forumInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
