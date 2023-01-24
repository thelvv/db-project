package thread

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

type ThreadInfo struct {
	ThreadApp app.ThreadAppInterface
	userApp   app.UserAppInterface
	logger    *zap.Logger
}

func NewThreadInfo(
	ThreadApp app.ThreadAppInterface,
	userApp app.UserAppInterface,
	logger *zap.Logger) *ThreadInfo {
	return &ThreadInfo{
		ThreadApp: ThreadApp,
		userApp:   userApp,
		logger:    logger,
	}
}

func (threadInfo *ThreadInfo) HandleCreateThread(w http.ResponseWriter, r *http.Request) {
	threadInfo.logger.Info("HandleCreateThread")
	vars := mux.Vars(r)
	slugOrID := vars[string(entity.SlugOrIDKey)]

	thread, err := threadInfo.ThreadApp.GetThreadForumAndID(slugOrID)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find post thread by id: %v", slugOrID),
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

	posts := make([]entity.Post, 0)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		threadInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(data, &posts)
	if err != nil {
		threadInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(posts) == 0 {
		body, err := json.Marshal(posts)
		if err != nil {
			threadInfo.logger.Info(
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

	for _, post := range posts {
		_, err = threadInfo.userApp.CheckIfUserExists(post.Author)
		if err != nil {
			msg := entity.Message{
				Text: fmt.Sprintf("Can't find post author by nickname: %v", post.Author),
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
	}

	err = threadInfo.ThreadApp.CreatePosts(thread, posts)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Parent post was created in another thread"),
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

	body, err := json.Marshal(posts)
	if err != nil {
		threadInfo.logger.Info(
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

func (threadInfo *ThreadInfo) HandleGetThreadDetails(w http.ResponseWriter, r *http.Request) {
	threadInfo.logger.Info("HandleGetThreadDetails")
	vars := mux.Vars(r)
	slugOrID := vars[string(entity.SlugOrIDKey)]

	threads, err := threadInfo.ThreadApp.GetThread(slugOrID)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find thread by slug: %v", slugOrID),
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

	body, err := json.Marshal(threads)
	if err != nil {
		threadInfo.logger.Info(
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

func (threadInfo *ThreadInfo) HandleUpdateThread(w http.ResponseWriter, r *http.Request) {
	threadInfo.logger.Info("HandleUpdateThread")
	vars := mux.Vars(r)
	slugOrID := vars[string(entity.SlugOrIDKey)]

	err := threadInfo.ThreadApp.CheckThread(slugOrID)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find thread by slug: %v", slugOrID),
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

	thread := &entity.Thread{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		threadInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(data, thread)
	if err != nil {
		threadInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if thread.Title == "" && thread.Message == "" {
		thread, err = threadInfo.ThreadApp.GetThread(slugOrID)
		if err != nil {
			threadInfo.logger.Info(
				err.Error(), zap.String("url", r.RequestURI),
				zap.String("method", r.Method))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		err = threadInfo.ThreadApp.UpdateThread(slugOrID, thread)
		if err != nil {
			threadInfo.logger.Info(
				err.Error(), zap.String("url", r.RequestURI),
				zap.String("method", r.Method))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	body, err := json.Marshal(thread)
	if err != nil {
		threadInfo.logger.Info(
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

func (threadInfo *ThreadInfo) HandleGetThreadPosts(w http.ResponseWriter, r *http.Request) {
	threadInfo.logger.Info("HandleGetThreadPosts")
	vars := mux.Vars(r)
	slugOrID := vars[string(entity.SlugOrIDKey)]

	err := threadInfo.ThreadApp.CheckThread(slugOrID)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find thread by slug: %v", slugOrID),
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
		threadInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sortParam, _ := queryParams[string(entity.SortKey)]
	sort := ""
	if sortParam == nil {
		sort = ""
	} else {
		sort = sortParam[0]
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

	posts, err := threadInfo.ThreadApp.GetThreadPosts(slugOrID, int32(limit), since, sort, desc)
	if err != nil {
		threadInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(posts)
	if err != nil {
		threadInfo.logger.Info(
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

func (threadInfo *ThreadInfo) HandleVoteForThread(w http.ResponseWriter, r *http.Request) {
	threadInfo.logger.Info("HandleVoteForThread")
	vars := mux.Vars(r)
	slugOrID := vars[string(entity.SlugOrIDKey)]

	vote := &entity.Vote{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		threadInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(data, vote)
	if err != nil {
		threadInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vote.Slug = slugOrID
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		id = 0
	}
	vote.ID = id

	thread, err := threadInfo.ThreadApp.VoteForThread(vote)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find thread by slug: %v", slugOrID),
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

	body, err := json.Marshal(thread)
	if err != nil {
		threadInfo.logger.Info(
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
