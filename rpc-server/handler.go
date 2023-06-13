package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	_ "github.com/go-sql-driver/mysql"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct{}

var db *sql.DB

const sendQuery = "INSERT INTO `messages` (`chat`, `text`, `sender`, `sendtime`) VALUES (?, ?, ?, ?)"

func init() {
	var err error
	db, err = sql.Open("mysql", "root:rootPasswordToBeUpdated@tcp(db:3306)/im")
	if err != nil {
		log.Fatalf("Error establishing connection to database: %s", err)
	}
}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	resp := rpc.NewSendResponse()
	if req.Message == nil {
		resp.Code, resp.Msg = consts.StatusBadRequest, "Chat, text and sender fields should not be empty"
		return resp, nil
	}

	if req.Message.Chat == "" || req.Message.Text == "" || req.Message.Sender == "" {
		resp.Code, resp.Msg = consts.StatusBadRequest, "Chat, text and sender fields should not be empty"
		return resp, nil
	}

	if strings.Count(req.Message.Chat, ":") != 1 {
		resp.Code, resp.Msg = consts.StatusBadRequest, "Chat field should contain only 1 : separating the 2 names, and names cannot contain :"
		return resp, nil
	}

	chatSplit := strings.Split(req.Message.Chat, ":")
	var senderIsValid = false
	for i := 0; i < len(chatSplit); i++ {
		if chatSplit[i] == req.Message.Sender {
			senderIsValid = true
			break
		} else if chatSplit[i] == "" {
			resp.Code, resp.Msg = consts.StatusBadRequest, "Names of people in chat field must not be empty"
			return resp, nil
		}
	}

	if !senderIsValid {
		resp.Code, resp.Msg = consts.StatusBadRequest, "Sender name should be one of the 2 names in the chat field"
		return resp, nil
	}

	_, err := db.ExecContext(context.Background(), sendQuery, req.Message.Chat, req.Message.Text, req.Message.Sender, time.Now().UnixNano())
	if err != nil {
		log.Printf("Error saving message in database: %s", err)
		resp.Code, resp.Msg = consts.StatusInternalServerError, "Error sending message"
		return resp, nil
	}

	resp.Code, resp.Msg = consts.StatusOK, "Message succcessfully sent"
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	var pullStatement = "SELECT * from `messages` WHERE `chat` = (?) AND `sendtime` >= (?) ORDER BY `sendtime` ASC"
	var cursor int64
	var hasMore bool
	var nextCursor int64
	var limit int32

	resp := rpc.NewPullResponse()

	if req.Chat == "" {
		resp.Code, resp.Msg = consts.StatusBadRequest, "Chat field should not be empty"
		return resp, nil
	}

	if strings.Count(req.Chat, ":") != 1 {
		resp.Code, resp.Msg = consts.StatusBadRequest, "Chat field should contain only 1 : separating the 2 names, and names cannot contain :"
		return resp, nil
	}

	chatSplit := strings.Split(req.Chat, ":")
	for i := 0; i < len(chatSplit); i++ {
		if chatSplit[i] == "" {
			resp.Code, resp.Msg = consts.StatusBadRequest, "Names of people in chat field must not be empty"
			return resp, nil
		}
	}

	if &req.Cursor != nil && req.Cursor > 0 {
		cursor = req.Cursor
	} else {
		cursor = 0
	}

	if &req.Limit != nil && req.Limit > 0 {
		pullStatement += " LIMIT " + fmt.Sprint(req.Limit+1)
		limit = req.Limit
	} else {
		pullStatement += " LIMIT 11"
		limit = 10
	}

	results, err := db.Query(pullStatement, req.Chat, cursor)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error retrieving messages from database: %s", err)
		resp.Code, resp.Msg = consts.StatusInternalServerError, "Error retrieving messages"
		return resp, nil
	}

	var currentIdx int32 = 0
	for results.Next() {
		if currentIdx < limit {
			var message rpc.Message
			var id int32
			err = results.Scan(&id, &message.Chat, &message.Text, &message.Sender, &message.SendTime)
			if err != nil {
				log.Printf("Error retrieving messages from database: %s", err)
				resp.Code, resp.Msg = consts.StatusInternalServerError, "Error retrieving messages"
				return resp, nil
			}
			resp.Messages = append(resp.Messages, &message)
		} else if currentIdx == limit {
			var id int32
			var chat string
			var text string
			var sender string
			err = results.Scan(&id, &chat, &text, &sender, &nextCursor)
			if err != nil {
				log.Printf("Error retrieving messages from database: %s", err)
				resp.Code, resp.Msg = consts.StatusInternalServerError, "Error retrieving messages"
				return resp, nil
			}
			hasMore = true
			resp.HasMore = &hasMore
			resp.NextCursor = &nextCursor
		}
		currentIdx += 1
	}

	if len(resp.Messages) > 0 && &req.Reverse != nil && *req.Reverse == true {
		for i, j := 0, len(resp.Messages)-1; i < j; i, j = i+1, j-1 {
			resp.Messages[i], resp.Messages[j] = resp.Messages[j], resp.Messages[i]
		}
	}

	resp.Code = consts.StatusOK
	return resp, nil
}
