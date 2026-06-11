package services

import (
	"bytes"
	"fmt"
	"medieval-store/models"

	"github.com/go-pdf/fpdf"
)

// InvoiceBundle groups everything needed to render one invoice page.
type InvoiceBundle struct {
	User       models.User
	Order      models.Order
	Items      []models.OrderItem
	ProductMap map[string]models.Product
}

// writeInvoicePage renders a single invoice onto a fresh page of the given PDF.
func writeInvoicePage(pdf *fpdf.Fpdf, user models.User, order models.Order, items []models.OrderItem, productMap map[string]models.Product) {
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Medieval Store - Official Invoice")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Invoice / Order #: %d", order.ID))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Customer: %s", user.Name))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Email: %s", user.Email))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Tax ID: %s", user.TaxID))
	pdf.Ln(8)

	pdf.Cell(40, 10, fmt.Sprintf("Date of Purchase: %s", order.CreatedAt.Format("January 02, 2006 15:04")))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Delivery Address: %s", order.DeliveryAddress))
	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Purchased Items:")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	total := 0.0
	for _, item := range items {
		name := "Unknown Artifact"
		price := item.Price
		if product, ok := productMap[item.ProductID]; ok {
			name = product.Name
		}

		line := fmt.Sprintf("%dx %s - $%.2f", item.Quantity, name, float64(item.Quantity)*price)
		pdf.Cell(40, 10, line)
		pdf.Ln(8)
		total += float64(item.Quantity) * price
	}

	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, fmt.Sprintf("Total Paid: $%.2f", total))
}

// GenerateInvoicePDF returns raw bytes []byte instead of saving to disk
func GenerateInvoicePDF(user models.User, order models.Order, items []models.OrderItem, productMap map[string]models.Product) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	writeInvoicePage(pdf, user, order, items, productMap)

	// Write the PDF directly to memory (RAM) instead of the hard drive
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

// GenerateBulkInvoicePDF renders every invoice in the slice into ONE document,
// one page per invoice — the sales manager's "save all invoices as PDF" export.
func GenerateBulkInvoicePDF(bundles []InvoiceBundle) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")

	if len(bundles) == 0 {
		// Still produce a valid (single-page) document stating there is nothing to export.
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, "No invoices found in the selected date range.")
	}

	for _, b := range bundles {
		writeInvoicePage(pdf, b.User, b.Order, b.Items, b.ProductMap)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}
