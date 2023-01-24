package post

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
	"strings"
)

type PostInfo struct {
	PostApp   app.PostAppInterface
	UserApp   app.UserAppInterface
	ThreadApp app.ThreadAppInterface
	ForumApp  app.ForumAppInterface
	logger    *zap.Logger
}

func NewPostInfo(
	PostApp app.PostAppInterface,
	UserApp app.UserAppInterface,
	ThreadApp app.ThreadAppInterface,
	ForumApp app.ForumAppInterface,
	logger *zap.Logger) *PostInfo {
	return &PostInfo{
		PostApp:   PostApp,
		UserApp:   UserApp,
		ThreadApp: ThreadApp,
		ForumApp:  ForumApp,
		logger:    logger,
	}
}

func (postInfo *PostInfo) HandleGetPostDetails(w http.ResponseWriter, r *http.Request) {
	postInfo.logger.Info("HandleGetPostDetails")
	vars := mux.Vars(r)
	idStr := vars[string(entity.IDKey)]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		postInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	post, err := postInfo.PostApp.GetPostDetails(id)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find post with id: %v", id),
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

	postInformation := entity.PostOutput{
		Post: post,
	}

	queryParams := r.URL.Query()

	relatedParam, _ := queryParams[string(entity.RelatedKey)]
	related := ""
	if relatedParam != nil {
		related = relatedParam[0]
	}
	if strings.Contains(related, "user") {
		author, err := postInfo.UserApp.GetUserByNickname(post.Author)
		if err != nil {
			msg := entity.Message{
				Text: fmt.Sprintf("Can't find user with id #%v\n", post.Author),
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
		postInformation.Author = author
	}

	if strings.Contains(related, "thread") {
		thread, err := postInfo.ThreadApp.GetThread(strconv.Itoa(post.Thread))
		if err != nil {
			msg := entity.Message{
				Text: fmt.Sprintf("Can't find thread forum by slug: %v", post.Thread),
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
		postInformation.Thread = thread
	}

	if strings.Contains(related, "forum") {
		forum, err := postInfo.ForumApp.GetForumDetails(post.Forum)
		if err != nil {
			msg := entity.Message{
				Text: fmt.Sprintf("Can't find forum by slug: %v", post.Forum),
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
		postInformation.Forum = forum
	}

	body, err := json.Marshal(postInformation)
	if err != nil {
		postInfo.logger.Info(
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

func (postInfo *PostInfo) HandleChangePost(w http.ResponseWriter, r *http.Request) {
	postInfo.logger.Info("starting PostDetailsUpdate")
	vars := mux.Vars(r)
	idStr := vars[string(entity.IDKey)]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		postInfo.logger.Info(err.Error(), zap.String("url", r.RequestURI), zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	post := &entity.Post{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		postInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(data, post)
	if err != nil {
		postInfo.logger.Info(
			err.Error(), zap.String("url", r.RequestURI),
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if post.Message == "" {
		post, err = postInfo.PostApp.GetPostDetails(id)
		if err != nil {
			msg := entity.Message{
				Text: fmt.Sprintf("Can't find post with id: %v", id),
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

		body, err := json.Marshal(post)
		if err != nil {
			postInfo.logger.Info(
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
	post.ID = id

	post, err = postInfo.PostApp.ChangePostMessage(post)
	if err != nil {
		msg := entity.Message{
			Text: fmt.Sprintf("Can't find post with id: %v", id),
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

	body, err := json.Marshal(post)
	if err != nil {
		postInfo.logger.Info(
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
