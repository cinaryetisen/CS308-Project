package services

import (
	"fmt"
	"medieval-store/models"

	"github.com/go-pdf/fpdf"
)

// GenerateInvoicePDF creates a pdf and returns the path
func GenerateInvoicePDF(user models.User, order models.Order, items []models.CartItem, productMap map[string]models.Product) (string, error) {
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
		//Default fallback values just in case a product got deleted from the database
		name := "Unknown Artifact"
		price := 0.00

		//Lookup the real name and price from the MongoDB data we passed in!
		if product, ok := productMap[item.ProductID]; ok {
			name = product.Name
			price = product.Price
		}

		//Example: "2x Iron Sword - $100.00"
		line := fmt.Sprintf("%dx %s - $%.2f", item.Quantity, name, float64(item.Quantity)*price)
		pdf.Cell(40, 10, line)
		pdf.Ln(8)

		total += float64(item.Quantity) * price
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
