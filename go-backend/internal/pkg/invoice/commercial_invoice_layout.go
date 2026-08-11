package invoice

import (
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
)

type pdfTextFunc func(string) string
type pdfFontFunc func(string, float64)

func drawAddressBox(pdf *fpdf.Fpdf, title string, address Address, text pdfTextFunc, setFont pdfFontFunc) {
	x, y := pdf.GetXY()
	width := 82.0
	lines := addressLines(address)
	height := 9.0 + float64(maxInt(len(lines), 1))*4.5 + 7

	pdf.SetDrawColor(214, 219, 226)
	pdf.SetFillColor(248, 249, 251)
	pdf.RoundedRect(x, y, width, height, 1.5, "1234", "DF")
	pdf.SetXY(x+4, y+3)
	setFont("B", 8)
	pdf.SetTextColor(70, 78, 88)
	pdf.CellFormat(width-8, 4, title, "", 1, "L", false, 0, "")
	setFont("", 8.5)
	pdf.SetTextColor(42, 48, 58)
	for _, line := range lines {
		pdf.SetX(x + 4)
		pdf.CellFormat(width-8, 4.5, text(line), "", 1, "L", false, 0, "")
	}
	pdf.SetXY(x, y+height+5)
}

func addressLines(address Address) []string {
	lines := make([]string, 0, 8)
	for _, value := range []string{
		address.Name,
		address.Company,
		address.Line1,
		address.Line2,
		joinNonEmpty(", ", address.City, address.State, address.PostalCode),
		address.Country,
		address.Email,
		address.Phone,
	} {
		if strings.TrimSpace(value) != "" {
			lines = append(lines, value)
		}
	}
	if len(lines) == 0 {
		return []string{"-"}
	}
	return lines
}

func drawTableHeader(pdf *fpdf.Fpdf, text pdfTextFunc, setFont pdfFontFunc) {
	widths := []float64{70, 27, 14, 31, 36}
	labels := []string{"Description", "SKU", "Qty", "Unit price", "Line total"}

	setFont("B", 8)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFillColor(42, 52, 64)
	for i, label := range labels {
		align := "L"
		if i >= 2 {
			align = "R"
		}
		pdf.CellFormat(widths[i], 7, text(label), "1", 0, align, true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetTextColor(42, 48, 58)
}

func itemRowHeight(pdf *fpdf.Fpdf, item LineItem, text pdfTextFunc) float64 {
	lines := pdf.SplitLines([]byte(text(textOrDash(item.Description))), 68)
	if len(lines) < 1 {
		return 8
	}
	return float64(len(lines))*4.5 + 3
}

func drawItemRow(pdf *fpdf.Fpdf, item LineItem, text pdfTextFunc, setFont pdfFontFunc) {
	widths := []float64{70, 27, 14, 31, 36}
	description := text(textOrDash(item.Description))
	descriptionLines := pdf.SplitLines([]byte(description), 68)
	rowHeight := itemRowHeight(pdf, item, text)
	y := pdf.GetY()

	pdf.SetDrawColor(220, 224, 230)
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(pdf.GetX(), y, sumFloat64(widths), rowHeight, "D")
	setFont("", 8)
	pdf.SetTextColor(42, 48, 58)
	pdf.MultiCell(widths[0], 4.5, strings.Join(byteLinesToStrings(descriptionLines), "\n"), "L", "L", false)

	pdf.SetXY(pdf.GetX()+widths[0], y)
	pdf.CellFormat(widths[1], rowHeight, text(textOrDash(item.SKU)), "L", 0, "L", false, 0, "")
	pdf.CellFormat(widths[2], rowHeight, fmt.Sprintf("%d", item.Quantity), "L", 0, "R", false, 0, "")
	pdf.CellFormat(widths[3], rowHeight, formatMoney(item.UnitPrice, ""), "L", 0, "R", false, 0, "")
	pdf.CellFormat(widths[4], rowHeight, formatMoney(item.Total, ""), "LR", 1, "R", false, 0, "")
	pdf.SetY(y + rowHeight)
}

func drawTotals(pdf *fpdf.Fpdf, document CommercialInvoice, text pdfTextFunc, setFont pdfFontFunc) {
	labelWidth := 124.0
	valueWidth := 54.0
	rows := []struct {
		label string
		value float64
	}{
		{"Subtotal", document.Subtotal},
		{"Shipping", document.Shipping},
		{"Tax", document.Tax},
		{"Discount", -document.Discount},
	}
	setFont("", 9)
	pdf.SetTextColor(70, 78, 88)
	for _, row := range rows {
		pdf.CellFormat(labelWidth, 5.5, text(row.label), "", 0, "R", false, 0, "")
		pdf.CellFormat(valueWidth, 5.5, text(formatMoney(row.value, document.Currency)), "", 1, "R", false, 0, "")
	}
	pdf.SetDrawColor(42, 52, 64)
	pdf.Line(labelWidth+16, pdf.GetY()+1, 194, pdf.GetY()+1)
	pdf.Ln(3)
	setFont("B", 11)
	pdf.SetTextColor(28, 36, 48)
	pdf.CellFormat(labelWidth, 7, "Total", "", 0, "R", false, 0, "")
	pdf.CellFormat(valueWidth, 7, text(formatMoney(document.Total, document.Currency)), "", 1, "R", false, 0, "")
}

func formatMoney(amount float64, currency string) string {
	value := fmt.Sprintf("%.2f", amount)
	if strings.TrimSpace(currency) == "" {
		return value
	}
	return strings.ToUpper(strings.TrimSpace(currency)) + " " + value
}

func joinNonEmpty(separator string, values ...string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			parts = append(parts, strings.TrimSpace(value))
		}
	}
	return strings.Join(parts, separator)
}

func sumFloat64(values []float64) float64 {
	var total float64
	for _, value := range values {
		total += value
	}
	return total
}

func byteLinesToStrings(lines [][]byte) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		result = append(result, string(line))
	}
	return result
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
