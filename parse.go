package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var reOrderType = regexp.MustCompile(`(成功购买|成功兑现|退票业务)`)
var reOrderNo = regexp.MustCompile(`订单号码([EB\d]+)`)
var reTicket = regexp.MustCompile(`(\d+)\.([^，]+)，([0-9]{4}年[0-9]{2}月[0-9]{2}日)([0-9]{2}:[0-9]{2})开，([^，]+)-([^，]+)，([A-Z0-9]+)次列车[，,]?([^，]+)，([^，]+)，(?:成人票)?[，,]?票价([0-9]+\.[0-9]+)元`)
var reRefund = regexp.MustCompile(`(：)([^，]+)，([0-9]{4}年[0-9]{2}月[0-9]{2}日)([0-9]{2}:[0-9]{2})开，([^，]+)-([^，]+)，([A-Z0-9]+)次列车[，,]?([^，]+)，([^，]+)，(?:成人票)?[，,]?票价([0-9]+\.[0-9]+)元[，,]?退票费([0-9]+\.[0-9]+)元，应退票款([0-9]+\.[0-9]+)元`)

type Ticket struct {
	// 订单号
	OrderNo string
	// 姓名
	Name string
	// 出发日期
	DepartDate string
	// 出发时间
	DepartTime string
	// 出发站
	FromStation string
	// 到达站
	ToStation string
	// 车次
	TrainNo string
	// 坐席号
	Seat string
	// 坐席类型
	SeatLevel string
	// 票价
	Price float64
	// 退票业务
	IsRefund bool
	// 退票费
	RefundFee float64
	// 应退票款
	RefundPrice float64
}

func parse(plainMail string) ([]Ticket, error) {
	var tickets []Ticket
	// 退票业务 候补购票业务 成功购买
	orderTypeMatch := reOrderType.FindStringSubmatch(plainMail)
	if len(orderTypeMatch) < 2 { // not found
		return tickets, fmt.Errorf("非指定购票类型(正常业务、候补业务、退票业务)")
	}
	orderType := orderTypeMatch[1]
	orderNoMatch := reOrderNo.FindStringSubmatch(plainMail)
	if len(orderNoMatch) < 2 { // not found
		return tickets, fmt.Errorf("未找到订单号")
	}
	orderNo := orderNoMatch[1]
	if orderType == "退票业务" {
		refundMatch := reRefund.FindStringSubmatch(plainMail)
		if len(refundMatch) != 13 {
			return tickets, fmt.Errorf("退票业务车票信息查找异常")
		}
		var t Ticket
		t.OrderNo = orderNo
		t.Name = refundMatch[2]
		t.DepartDate = refundMatch[3]
		t.DepartTime = refundMatch[4]
		t.FromStation = refundMatch[5]
		t.ToStation = refundMatch[6]
		t.TrainNo = refundMatch[7]
		t.Seat = refundMatch[8]
		t.SeatLevel = refundMatch[9]
		t.Price = parsePrice(refundMatch[10])
		t.IsRefund = true
		t.RefundFee = parsePrice(refundMatch[11])
		t.RefundPrice = parsePrice(refundMatch[12])
		tickets = append(tickets, t)
	} else {
		ticketMatch := reTicket.FindAllStringSubmatch(plainMail, -1)
		if len(ticketMatch) == 0 {
			return tickets, fmt.Errorf("未找到车票信息")
		}
		for _, ticket := range ticketMatch {
			if len(ticket) != 11 {
				return tickets, fmt.Errorf("车票信息查找异常")
			}
			var t Ticket
			t.OrderNo = orderNo
			t.Name = ticket[2]
			t.DepartDate = ticket[3]
			t.DepartTime = ticket[4]
			t.FromStation = ticket[5]
			t.ToStation = ticket[6]
			t.TrainNo = ticket[7]
			t.Seat = ticket[8]
			t.SeatLevel = ticket[9]
			t.Price = parsePrice(ticket[10])
			tickets = append(tickets, t)
		}
	}
	return tickets, nil
}

func parsePrice(price string) float64 {
	price = strings.TrimSpace(price)
	if len(price) == 0 {
		return 0.0
	}
	p, _ := strconv.ParseFloat(price, 64)
	return p
}

func parseDepartDatetime(departDate string, departTime string) (time.Time, error) {
	return time.Parse("2006年01月02日 15:04", departDate+" "+departTime)
}
