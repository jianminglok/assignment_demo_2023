package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
	db, err = sql.Open("mysql", "root:mauFJcuf5dhRMQrjj@tcp(db:3306)/im")
	if err != nil {
		log.Fatalf("Error establishing connection to database: %s", err)
	}
}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	_, err := db.ExecContext(context.Background(), sendQuery, req.Message.Chat, req.Message.Text, req.Message.Sender, time.Now().UnixNano())
	resp := rpc.NewSendResponse()
	if err != nil {
		log.Fatalf("Error saving message in database: %s", err)
		resp.Code, resp.Msg = consts.StatusInternalServerError, "Error sending message"
		return resp, nil
	}
	resp.Code, resp.Msg = consts.StatusOK, "Message succcessfully sent"
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	var pullStatement = "SELECT * from `messages` WHERE `chat` = (?) AND `sendtime` >= (?) ORDER BY `sendtime` ASC"
	var countPullStatement string
	var cursor int64
	if &req.Cursor != nil && req.Cursor > 0 {
		cursor = req.Cursor
	} else {
		cursor = 0
	}
	countPullStatement = "SELECT COUNT(*) from (SELECT * from `messages` WHERE `chat` = (?) AND `sendtime` >= (?) ORDER BY `sendtime` ASC) AS count"
	if &req.Limit != nil && req.Limit > 0 {
		pullStatement += " LIMIT " + fmt.Sprint(req.Limit)
	} else {
		pullStatement += " LIMIT 10"
	}
	results, err := db.Query(pullStatement, req.Chat, cursor)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	}
	var resultsCount int32
	err = db.QueryRow(countPullStatement, req.Chat, cursor).Scan(&resultsCount)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	}
	resp := rpc.NewPullResponse()
	for results.Next() {
		var message rpc.Message
		var id int32
		err = results.Scan(&id, &message.Chat, &message.Text, &message.Sender, &message.SendTime)
		if err != nil {
			log.Fatal(err)
		}
		resp.Messages = append(resp.Messages, &message)
	}
	if &req.Reverse != nil && *req.Reverse == true {
		for i, j := 0, len(resp.Messages)-1; i < j; i, j = i+1, j-1 {
			resp.Messages[i], resp.Messages[j] = resp.Messages[j], resp.Messages[i]
		}
	}
	if err != nil {
		log.Fatalf("Error retrieving messages from database: %s", err)
		resp.Code, resp.Msg = consts.StatusInternalServerError, "Error pulling messages"
		return resp, nil
	}
	resp.Code = consts.StatusOK
	var hasMore bool
	var nextCursor int64
	if &req.Reverse != nil && *req.Reverse == true {
		var testStatement = "SELECT sendtime from `messages` WHERE `chat` = (?) AND `sendtime` < (?) ORDER BY `sendtime` DESC"
		if &req.Limit != nil && req.Limit > 0 {
			testStatement += " LIMIT " + fmt.Sprint(req.Limit)
		} else {
			pullStatement += " LIMIT 10"
		}
		resultsOther, err := db.Query(testStatement, req.Chat, resp.Messages[len(resp.Messages)-1].SendTime)
		if err != nil {
			log.Println(err)
		} else {
			for resultsOther.Next() {
				err := resultsOther.Scan(&nextCursor)
				if err != nil {
					if err == sql.ErrNoRows {
						hasMore = false
					}
					log.Fatal(err)
				} else {
					hasMore = true
					resp.NextCursor = &nextCursor
				}
			}
		}

	} else {
		result := db.QueryRow("SELECT sendtime from `messages` WHERE `chat` = (?) AND `sendtime` > (?) ORDER BY `sendtime` ASC", req.Chat, resp.Messages[len(resp.Messages)-1].SendTime)
		if result != nil {
			err := result.Scan(&nextCursor)
			if err != nil {
				if err == sql.ErrNoRows {
					hasMore = false
				}
				log.Println(err)
			} else {
				hasMore = true
				resp.NextCursor = &nextCursor
			}
		}
	}
	resp.HasMore = &hasMore
	return resp, nil
}
