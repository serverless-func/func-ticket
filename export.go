package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"slices"
	"strings"
)

const sheetName = "Sheet1"

var sheetHeaders = []string{"订单号", "姓名", "出发日期", "出发时间", "出发站", "到达站", "车次", "座位号", "座位类型", "票价", "退票业务", "退票费", "应退票款"}

func export(tickets []Ticket) (*excelize.File, error) {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)

	style, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	lastHeader := fmt.Sprintf("%s1", string(rune('A'+len(sheetHeaders)-1)))
	_ = f.SetCellStyle(sheetName, "A1", lastHeader, style)
	for i, header := range sheetHeaders {
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+i)), 1), header)
		colName, _ := excelize.ColumnNumberToName(i + 1)
		// 根据表头长度设置列宽（可自定义逻辑）
		_ = f.SetColWidth(sheetName, colName, colName, float64(len(header)*2))
	}
	// 用 slices.SortedFunc 按时间排序
	slices.SortFunc(tickets, func(a, b Ticket) int {
		return strings.Compare(a.DepartDate, b.DepartDate)
	})
	for i, ticket := range tickets {
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+0), i+2), ticket.OrderNo)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+1), i+2), ticket.Name)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+2), i+2), ticket.DepartDate)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+3), i+2), ticket.DepartTime)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+4), i+2), ticket.FromStation)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+5), i+2), ticket.ToStation)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+6), i+2), ticket.TrainNo)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+7), i+2), ticket.Seat)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+8), i+2), ticket.SeatLevel)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+9), i+2), ticket.Price)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+10), i+2), ticket.IsRefund)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+11), i+2), ticket.RefundFee)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string('A'+12), i+2), ticket.RefundPrice)
	}
	return f, nil
}
