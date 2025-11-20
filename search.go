package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
	"slices"
	"strings"
	"time"
)

const maxSearchDays = 30
const filterSubject = "网上购票系统"

func search(cfg searchConfig) ([]string, error) {
	var mails []string
	c, err := client.DialTLS("imap.exmail.qq.com:993", nil)
	if err != nil {
		return mails, fmt.Errorf("dial imap server error: %s", err.Error())
	}
	log.Println("dial ...")
	defer func() {
		_ = c.Logout()
	}()

	if err := c.Login(cfg.Username, cfg.Password); err != nil {
		return mails, fmt.Errorf("email login error: %s", err.Error())
	}
	log.Println("login ...")
	_, err = c.Select("其他文件夹/邮件转移", true)
	if err != nil {
		return mails, fmt.Errorf("open inbox error: %s", err.Error())
	}
	log.Println("select folder ...")
	var allUID []uint32
	// cfg.Days 按 maxSearchDays 分组查询
	splitDays(cfg.Days, time.Now(), func(start, end time.Time, index int) {
		uidList := searchTimeRage(c, start, end)
		allUID = append(allUID, uidList...)
	})
	log.Println("search by time done")
	if len(allUID) == 0 {
		log.Println("no new message")
		return mails, nil
	}
	chunks := slices.Chunk(allUID, 30)
	total := len(allUID)
	count := 0
	for chunk := range chunks {
		log.Printf("fetch (%d-%d) / %d ...", count, count+len(chunk), total)
		mails = append(mails, fetchMatchedMail(c, chunk)...)
		count += len(chunk)
	}
	return mails, nil
}

func splitDays(days int, endTime time.Time, callback func(start, end time.Time, index int)) {
	if days <= 0 {
		fmt.Println("总天数必须大于0")
		return
	}

	totalDays := days
	currentEnd := endTime
	index := 0

	// 循环拆分区间，直到总天数用完
	for totalDays > 0 {
		// 计算当前区间的天数（最后一个区间可能不足30天）
		currentIntervalDays := maxSearchDays
		if totalDays < maxSearchDays {
			currentIntervalDays = totalDays
		}

		// 计算当前区间的开始时间（结束时间 - 当前区间天数）
		currentStart := currentEnd.AddDate(0, 0, -currentIntervalDays)

		// 执行回调，处理当前区间（如搜索该时间段的邮件）
		callback(currentStart, currentEnd, index)

		// 更新下一轮循环的参数
		currentEnd = currentStart        // 下一个区间的结束时间 = 当前区间的开始时间
		totalDays -= currentIntervalDays // 剩余未拆分的天数
		index++
	}
}

// 搜索时间范围 (Since, Since + 30)
func searchTimeRage(c *client.Client, start, end time.Time) []uint32 {
	filter := imap.NewSearchCriteria()
	filter.Since = start
	filter.Before = end
	log.Printf("search by time %s ~ %s ...", filter.Since.Format("2006-01-02"), filter.Before.Format("2006-01-02"))
	uidList, err := c.UidSearch(filter)
	if err == nil {
		uidList := searchSubject(c, filterSubject, uidList)
		log.Printf("search by time %s ~ %s ... (%d)", filter.Since.Format("2006-01-02"), filter.Before.Format("2006-01-02"), len(uidList))
		return uidList
	}
	return nil
}

func searchSubject(c *client.Client, filterSubject string, allUID []uint32) []uint32 {
	uidSet := new(imap.SeqSet)
	uidSet.AddNum(allUID...)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(uidSet, []imap.FetchItem{imap.FetchUid, imap.FetchEnvelope}, messages)
	}()
	var matchedUIDs []uint32
	for msg := range messages {
		subject := msg.Envelope.Subject
		if strings.HasPrefix(subject, "=?GBK?B?") {
			subject = strings.TrimPrefix(subject, "=?GBK?B?")
			subject = strings.TrimSuffix(subject, "?=")
			decoded, err := base64.StdEncoding.DecodeString(subject)
			if err == nil {
				reader := transform.NewReader(bytes.NewReader(decoded), simplifiedchinese.GBK.NewDecoder())
				utf8Bytes, err := io.ReadAll(reader)
				if err == nil {
					subject = string(utf8Bytes)
				}
			}
		} else if strings.HasPrefix(subject, "=?UTF-8?B?") {
			subject = strings.TrimPrefix(subject, "=?UTF-8?B?")
			subject = strings.TrimSuffix(subject, "?=")
			decoded, err := base64.StdEncoding.DecodeString(subject)
			if err == nil {
				subject = string(decoded)
			}
		}
		if strings.Contains(subject, filterSubject) {
			matchedUIDs = append(matchedUIDs, msg.Uid)
		}
	}
	return matchedUIDs
}

func fetchMatchedMail(c *client.Client, matchedUID []uint32) []string {
	matchedSet := new(imap.SeqSet)
	matchedSet.AddNum(matchedUID...)

	section := &imap.BodySectionName{}
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(matchedSet, []imap.FetchItem{section.FetchItem()}, messages)
	}()
	mails := make([]string, 0)
	for msg := range messages {
		r := msg.GetBody(section)
		if r == nil {
			log.Println("Server didn't returned message body")
			continue
		}
		mr, err := mail.ReadMessage(r)
		if err != nil {
			log.Printf("read email error: %s\n", err.Error())
			continue
		}
		mediaType, params, _ := mime.ParseMediaType(mr.Header.Get("Content-Type"))
		if strings.HasPrefix(mediaType, "multipart/") {
			br := multipart.NewReader(mr.Body, params["boundary"])
			for {
				p, err := br.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("email body parse error: %s\n", err.Error())
					break
				}

				slurp, err := io.ReadAll(p)
				if err != nil {
					log.Printf("email body parse error: %s\n", err.Error())
					break
				}
				var reader io.Reader
				if p.Header.Get("Content-Type") == "text/html; charset=gbk" {
					reader = transform.NewReader(bytes.NewReader(slurp), simplifiedchinese.GBK.NewDecoder())
				} else if p.Header.Get("Content-Type") == "text/html; charset=utf-8" {
					reader = bytes.NewReader(slurp)
				}

				doc, err := goquery.NewDocumentFromReader(reader)
				if err != nil {
					log.Printf("parse html error: %s\n", err.Error())
					break
				}
				mails = append(mails, strings.Map(func(r rune) rune {
					if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
						return -1 // 移除该字符
					}
					return r
				}, doc.Text()))
			}
		}
	}
	return mails
}
