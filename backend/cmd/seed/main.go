package main

import (
	"context"
	"log"
	"os"
	"time"

	"medieval-store/config"
	"medieval-store/models"
	"medieval-store/security"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	// 1. Load the .env file from the root directory
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or couldn't load it")
	}

	// 2. Connect to BOTH Databases
	if err := config.ConnectMongo(); err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	if err := config.ConnectPostgres(); err != nil {
		log.Fatalf("PostgreSQL connection failed: %v", err)
	}

	// ==========================================
	// PART A: SEED POSTGRESQL (USERS & MANAGERS)
	// ==========================================
	log.Println("Migrating and Seeding PostgreSQL Users...")

	// Ensure the users table exists before inserting
	config.DB.AutoMigrate(&models.User{})

	// Generate a secure password hash using YOUR security package
	defaultPassword, err := security.HashPassword("password123")
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Define our core users for Sprint 5
	users := []models.User{
		{
			Name:        "Sales Manager Admin",
			Email:       "sales@medievalstore.com",
			Password:    defaultPassword,
			Role:        "sales_manager",
			TaxID:       "TX-999-SALES",
			HomeAddress: "123 Coinshire Vaults, Market District",
		},
		{
			Name:        "Product Manager Admin",
			Email:       "product@medievalstore.com",
			Password:    defaultPassword,
			Role:        "product_manager",
			TaxID:       "TX-888-PROD",
			HomeAddress: "456 Forge Lane, Artisan District",
		},
		{
			Name:        "Test Customer",
			Email:       "customer@medievalstore.com",
			Password:    defaultPassword,
			Role:        "customer",
			TaxID:       "TX-111-CUST",
			HomeAddress: "789 Peasant Way, Lower Ward",
		},
	}

	// Insert users. FirstOrCreate prevents errors if you run this script multiple times!
	for _, user := range users {
		// Note: Your BeforeSave hook in user.go will automatically encrypt the TaxID and HomeAddress here!
		if err := config.DB.Where(models.User{Email: user.Email}).FirstOrCreate(&user).Error; err != nil {
			log.Printf("Failed to seed user %s: %v\n", user.Email, err)
		}
	}
	log.Println("Successfully seeded Manager and Customer accounts!")

	// ==========================================
	// PART B: SEED MONGODB (PRODUCTS)
	// ==========================================
	log.Println("Clearing and Seeding MongoDB Products...")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	imgBase := os.Getenv("IMAGE_BASE_URL")
	if imgBase == "" {
		imgBase = "http://localhost:8080/images"
	}

	// Clear existing products
	if _, err := collection.DeleteMany(ctx, bson.M{}); err != nil {
		log.Fatalf("Failed to clear existing products: %v", err)
	}

	// Create the Mock Data with COST included (Cost = Price * 0.6)
	mockProducts := []interface{}{
		// --- SPRINT 1 ORIGINAL WEAPONS & SPELLS ---
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Broadsword of the Bear",
			Model:        "SWD-001",
			SerialNumber: "SN-99812",
			Description:  "A heavy, two-handed broadsword forged from high-carbon steel. Excellent for close-quarters combat.",
			Quantity:     15,
			Cost:         150.00, // 60% of 250
			Price:        250.00,
			Discount:     0,
			Warranty:     "Lifetime (Against dragon fire)",
			Distributor:  "Ironhelm Forge",
			Category:     "Weapons",
			ImageURL:     imgBase + "/broadsword-of-the-bear.png",
			Tags:         []string{"sword", "melee", "heavy"},
			Rating:       4.8,
			ReviewCount:  24,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Scroll of Lesser Healing",
			Model:        "SPL-104",
			SerialNumber: "SN-11002",
			Description:  "A parchment inscribed with ancient runes. Instantly heals minor cuts and bruises when read aloud.",
			Quantity:     100,
			Cost:         9.30, // 60% of 15.50
			Price:        15.50,
			Discount:     10.0,
			Warranty:     "No refunds once unsealed",
			Distributor:  "The Arcane Order",
			Category:     "Spells",
			ImageURL:     imgBase + "/scroll-of-lesser-healing.png",
			Tags:         []string{"magic", "healing", "consumable"},
			Rating:       4.2,
			ReviewCount:  156,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Elven Cloak of Hiding",
			Model:        "AMR-045",
			SerialNumber: "SN-44332",
			Description:  "A lightweight cloak that shifts color to match your surroundings. Currently out of stock due to high demand.",
			Quantity:     0,
			Cost:         300.00, // 60% of 500
			Price:        500.00,
			Discount:     0,
			Warranty:     "1 Year",
			Distributor:  "Rivendell Textiles",
			Category:     "Apparel",
			ImageURL:     imgBase + "/elven-cloak-of-hiding.png",
			Tags:         []string{"stealth", "armor", "rare"},
			Rating:       5.0,
			ReviewCount:  8,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// --- WEAPONS ---
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Ranger's Longbow",
			Model:        "BOW-088",
			SerialNumber: "SN-77210",
			Description:  "Crafted from flexible yew wood. Includes a quiver of 20 iron-tipped arrows.",
			Quantity:     45,
			Cost:         72.00, // 60% of 120
			Price:        120.00,
			Discount:     0,
			Warranty:     "6 Months",
			Distributor:  "Woodland Fletchers",
			Category:     "Weapons",
			ImageURL:     imgBase + "/rangers-longbow.png",
			Tags:         []string{"bow", "ranged", "wood"},
			Rating:       4.6,
			ReviewCount:  89,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Obsidian Dagger",
			Model:        "DAG-012",
			SerialNumber: "SN-99100",
			Description:  "A sleek, razor-sharp dagger forged from volcanic glass. Perfect for rogues.",
			Quantity:     8,
			Cost:         210.00, // 60% of 350
			Price:        350.00,
			Discount:     5.0,
			Warranty:     "No Warranty",
			Distributor:  "Shadow Traders",
			Category:     "Weapons",
			ImageURL:     imgBase + "/obsidian-dagger.png",
			Tags:         []string{"dagger", "stealth", "melee"},
			Rating:       4.9,
			ReviewCount:  12,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Heavy Halberd",
			Model:        "HLB-044",
			SerialNumber: "SN-33112",
			Description:  "A versatile polearm featuring an axe blade topped with a spike.",
			Quantity:     20,
			Cost:         108.00, // 60% of 180
			Price:        180.00,
			Discount:     0,
			Warranty:     "2 Years",
			Distributor:  "Ironhelm Forge",
			Category:     "Weapons",
			ImageURL:     imgBase + "/heavy-halberd.png",
			Tags:         []string{"polearm", "heavy", "melee"},
			Rating:       4.1,
			ReviewCount:  34,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Dwarven Crossbow",
			Model:        "CBW-001",
			SerialNumber: "SN-88221",
			Description:  "A mechanical marvel that fires bolts with devastating piercing power.",
			Quantity:     12,
			Cost:         240.00, // 60% of 400
			Price:        400.00,
			Discount:     15.0,
			Warranty:     "5 Years",
			Distributor:  "Deep Mountain Guild",
			Category:     "Weapons",
			ImageURL:     imgBase + "/dwarven-crossbow.png",
			Tags:         []string{"ranged", "mechanical", "dwarven"},
			Rating:       4.7,
			ReviewCount:  56,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// --- APPAREL & ARMOR ---
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Standard Iron Chainmail",
			Model:        "AMR-101",
			SerialNumber: "SN-55441",
			Description:  "Basic but reliable protection against slashing attacks. Weighs 40 lbs.",
			Quantity:     30,
			Cost:         120.00, // 60% of 200
			Price:        200.00,
			Discount:     0,
			Warranty:     "1 Year",
			Distributor:  "Ironhelm Forge",
			Category:     "Apparel",
			ImageURL:     imgBase + "/standard-iron-chainmail.png",
			Tags:         []string{"armor", "iron", "heavy"},
			Rating:       3.8,
			ReviewCount:  112,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Mithril Chestplate",
			Model:        "AMR-999",
			SerialNumber: "SN-00001",
			Description:  "Incredibly rare armor that is as light as a feather but harder than dragon scales.",
			Quantity:     1,
			Cost:         3000.00, // 60% of 5000
			Price:        5000.00,
			Discount:     0,
			Warranty:     "Lifetime",
			Distributor:  "Deep Mountain Guild",
			Category:     "Apparel",
			ImageURL:     imgBase + "/mithril-chestplate.png",
			Tags:         []string{"armor", "legendary", "lightweight"},
			Rating:       5.0,
			ReviewCount:  2,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Hardened Leather Bracers",
			Model:        "AMR-022",
			SerialNumber: "SN-11223",
			Description:  "Provides decent wrist and forearm protection without sacrificing mobility.",
			Quantity:     85,
			Cost:         27.00, // 60% of 45
			Price:        45.00,
			Discount:     0,
			Warranty:     "30 Days",
			Distributor:  "Woodland Fletchers",
			Category:     "Apparel",
			ImageURL:     imgBase + "/hardened-leather-bracers.png",
			Tags:         []string{"armor", "leather", "accessories"},
			Rating:       4.3,
			ReviewCount:  67,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Steel Kite Shield",
			Model:        "SHD-005",
			SerialNumber: "SN-77665",
			Description:  "A large, teardrop-shaped shield ideal for deflecting arrows and cavalry charges.",
			Quantity:     0,
			Cost:         90.00, // 60% of 150
			Price:        150.00,
			Discount:     0,
			Warranty:     "1 Year",
			Distributor:  "Ironhelm Forge",
			Category:     "Apparel",
			ImageURL:     imgBase + "/steel-kite-shield.png",
			Tags:         []string{"shield", "defense", "steel"},
			Rating:       4.5,
			ReviewCount:  41,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// --- SPELLS & MAGIC ---
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Potion of Invisibility",
			Model:        "PTN-007",
			SerialNumber: "SN-11009",
			Description:  "Grants the drinker complete unseen status for exactly 3 minutes. Tastes like mint.",
			Quantity:     50,
			Cost:         45.00, // 60% of 75
			Price:        75.00,
			Discount:     0,
			Warranty:     "No Refunds",
			Distributor:  "The Arcane Order",
			Category:     "Spells",
			ImageURL:     imgBase + "/potion-of-invisibility.png",
			Tags:         []string{"potion", "stealth", "magic"},
			Rating:       4.8,
			ReviewCount:  203,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Wand of Fireballs",
			Model:        "WND-332",
			SerialNumber: "SN-99887",
			Description:  "A mahogany wand core-infused with a phoenix feather. Contains 50 charges.",
			Quantity:     5,
			Cost:         510.00, // 60% of 850
			Price:        850.00,
			Discount:     50.0,
			Warranty:     "Void if exposed to water",
			Distributor:  "The Arcane Order",
			Category:     "Spells",
			ImageURL:     imgBase + "/wand-of-fireballs.png",
			Tags:         []string{"magic", "fire", "ranged"},
			Rating:       4.2,
			ReviewCount:  19,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Minor Mana Potion",
			Model:        "PTN-002",
			SerialNumber: "SN-11011",
			Description:  "Restores a small amount of magical energy. Essential for apprentice mages.",
			Quantity:     200,
			Cost:         6.00, // 60% of 10
			Price:        10.00,
			Discount:     0,
			Warranty:     "No Refunds",
			Distributor:  "The Arcane Order",
			Category:     "Spells",
			ImageURL:     imgBase + "/minor-mana-potion.png",
			Tags:         []string{"potion", "mana", "consumable"},
			Rating:       4.0,
			ReviewCount:  412,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Tome of Ancient Lore",
			Model:        "BOK-001",
			SerialNumber: "SN-55667",
			Description:  "A dusty, leather-bound book containing forgotten history and basic enchantments.",
			Quantity:     3,
			Cost:         180.00, // 60% of 300
			Price:        300.00,
			Discount:     0,
			Warranty:     "As-Is Condition",
			Distributor:  "Silvermoon Library",
			Category:     "Spells",
			ImageURL:     imgBase + "/tome-of-ancient-lore.png",
			Tags:         []string{"book", "lore", "magic"},
			Rating:       4.9,
			ReviewCount:  6,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// --- ACCESSORIES & MISC ---
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Ring of Vitality",
			Model:        "RNG-010",
			SerialNumber: "SN-22334",
			Description:  "A silver band set with a ruby. Slightly increases the wearer's stamina.",
			Quantity:     14,
			Cost:         132.00, // 60% of 220
			Price:        220.00,
			Discount:     20.0,
			Warranty:     "Lifetime",
			Distributor:  "Shadow Traders",
			Category:     "Accessories",
			ImageURL:     imgBase + "/ring-of-vitality.png",
			Tags:         []string{"jewelry", "enchanted", "health"},
			Rating:       4.6,
			ReviewCount:  45,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Crystal Ball of Scrying",
			Model:        "MIS-055",
			SerialNumber: "SN-33445",
			Description:  "Allows the user to glimpse faraway places. Requires intense concentration.",
			Quantity:     2,
			Cost:         720.00, // 60% of 1200
			Price:        1200.00,
			Discount:     0,
			Warranty:     "No Warranty (Fragile)",
			Distributor:  "The Arcane Order",
			Category:     "Accessories",
			ImageURL:     imgBase + "/crystal-ball-of-scrying.png",
			Tags:         []string{"magic", "divination", "tool"},
			Rating:       3.9,
			ReviewCount:  14,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Amulet of Protection",
			Model:        "AML-003",
			SerialNumber: "SN-77889",
			Description:  "Wards off minor curses and hexes. Glows faintly when danger is near.",
			Quantity:     0,
			Cost:         270.00, // 60% of 450
			Price:        450.00,
			Discount:     0,
			Warranty:     "5 Years",
			Distributor:  "Silvermoon Library",
			Category:     "Accessories",
			ImageURL:     imgBase + "/amulet-of-protection.png",
			Tags:         []string{"jewelry", "defense", "magic"},
			Rating:       4.7,
			ReviewCount:  28,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Lockpick Set",
			Model:        "MIS-009",
			SerialNumber: "SN-99001",
			Description:  "A 10-piece set of steel tools for opening standard chest and door locks.",
			Quantity:     40,
			Cost:         21.00, // 60% of 35
			Price:        35.00,
			Discount:     0,
			Warranty:     "30 Days",
			Distributor:  "Shadow Traders",
			Category:     "Accessories",
			ImageURL:     imgBase + "/lockpick-set.png",
			Tags:         []string{"tools", "stealth", "rogue"},
			Rating:       4.4,
			ReviewCount:  132,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		models.Product{
			ID:           primitive.NewObjectID(),
			Name:         "Torch",
			Model:        "MIS-001",
			SerialNumber: "SN-11111",
			Description:  "A simple wooden stick wrapped in pitch-soaked rags. Burns for 1 hour.",
			Quantity:     500,
			Cost:         1.20, // 60% of 2
			Price:        2.00,
			Discount:     0,
			Warranty:     "No Warranty",
			Distributor:  "Woodland Fletchers",
			Category:     "Accessories",
			ImageURL:     imgBase + "/torch.png",
			Tags:         []string{"tool", "light", "consumable"},
			Rating:       4.1,
			ReviewCount:  890,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	result, err := collection.InsertMany(ctx, mockProducts)
	if err != nil {
		log.Fatalf("Failed to insert mock data: %v", err)
	}

	log.Printf("Successfully seeded %d medieval products into MongoDB!\n", len(result.InsertedIDs))

	// ==========================================
	// PART C: SEED MONGODB CATEGORIES
	// ==========================================
	categoriesCollection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	if _, err := categoriesCollection.DeleteMany(ctx, bson.M{}); err != nil {
		log.Fatalf("Failed to clear existing categories: %v", err)
	}
	log.Println("Cleared old category data...")

	now := time.Now()
	mockCategories := []interface{}{
		models.Category{Name: "Accessories", CreatedAt: now},
		models.Category{Name: "Apparel", CreatedAt: now},
		models.Category{Name: "Spells", CreatedAt: now},
		models.Category{Name: "Weapons", CreatedAt: now},
	}

	catResult, err := categoriesCollection.InsertMany(ctx, mockCategories)
	if err != nil {
		log.Fatalf("Failed to insert mock categories: %v", err)
	}
	log.Printf("Successfully seeded %d categories into MongoDB!\n", len(catResult.InsertedIDs))
}
