package services

import (
	"fmt"
	"github.com/go-pdf/fpdf"
	"medieval-store/models"
)

// Function to create a pdf and return the path
func GenerateInvoicePDF(user models.User, order models.Order, items []models.CartItem) (string, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	//Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Medieval Store - Official Invoice")
	pdf.Ln(12)

	//Customer Info
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Customer: %s", user.Name))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Email: %s", user.Email))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Tax ID: %s", user.TaxID))
	pdf.Ln(15)

	//Order Items
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Purchased Items:")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	total := 0.0
	for _, item := range items {
		//Example: "2x Iron Sword - $100"
		line := fmt.Sprintf("%dx %s - $%.2f", item.Quantity, item.ProductID, float64(item.Quantity)*10.0) // Replace 10.0 with actual price
		pdf.Cell(40, 10, line)
		pdf.Ln(8)
		total += float64(item.Quantity) * 10.0
	}

	//Total
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, fmt.Sprintf("Total Paid: $%.2f", total))

	//Save to a temporary file
	fileName := fmt.Sprintf("invoice_order_%d.pdf", order.ID)
	err := pdf.OutputFileAndClose(fileName)
	return fileName, err
}
