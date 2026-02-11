package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rohitaj002/product-inventory-CLI/internal/domain"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func init() {
	// Create Command
	createCmd.Flags().String("name", "", "Product name")
	createCmd.Flags().Float64("price", 0, "Product price")
	createCmd.Flags().Int("quantity", 0, "Product quantity")
	createCmd.Flags().String("category", "", "Product category")
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("price")
	createCmd.MarkFlagRequired("quantity")
	rootCmd.AddCommand(createCmd)

	// Get Command
	rootCmd.AddCommand(getCmd)

	// List Command
	listCmd.Flags().String("category", "", "Filter by category")
	listCmd.Flags().Float64("min-price", 0, "Minimum price")
	listCmd.Flags().Float64("max-price", 0, "Maximum price")
	listCmd.Flags().Bool("json", false, "Output in JSON format")            // --json flag
	listCmd.Flags().String("output", "table", "Output format (table|json)") // --output flag overrides --json if set?
	// Supports table (default) and json output format.
	rootCmd.AddCommand(listCmd)

	updateCmd.Flags().String("name", "", "New product name")
	updateCmd.Flags().Float64("price", -1, "New product price")
	updateCmd.Flags().Int("quantity", -1, "New product quantity")
	updateCmd.Flags().String("category", "", "New product category")
	rootCmd.AddCommand(updateCmd)

	// Delete Command
	deleteCmd.Flags().Bool("force", false, "Skip confirmation")
	rootCmd.AddCommand(deleteCmd)

	// Import Command
	importCmd.Flags().String("file", "", "JSON file to import")
	importCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(importCmd)

	// Export Command
	exportCmd.Flags().String("file", "export.json", "File to export to")
	exportCmd.Flags().String("category", "", "Filter by category")
	rootCmd.AddCommand(exportCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new product",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		price, _ := cmd.Flags().GetFloat64("price")
		quantity, _ := cmd.Flags().GetInt("quantity")
		category, _ := cmd.Flags().GetString("category")

		if price < 0 {
			return fmt.Errorf("price cannot be negative")
		}

		product := domain.Product{
			ID:       uuid.New().String(),
			Name:     name,
			Price:    price,
			Quantity: quantity,
			Category: category,
		}

		if err := appStore.Create(cmd.Context(), product); err != nil {
			return err
		}

		fmt.Printf("Product created successfully: %s\n", product.ID)
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get a product by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		product, err := appStore.Get(cmd.Context(), id)
		if err != nil {
			return err
		}

		printTable([]domain.Product{product})
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all products",
	RunE: func(cmd *cobra.Command, args []string) error {
		category, _ := cmd.Flags().GetString("category")
		minPrice, _ := cmd.Flags().GetFloat64("min-price")
		maxPrice, _ := cmd.Flags().GetFloat64("max-price")
		output, _ := cmd.Flags().GetString("output")

		filter := domain.ListFilter{}
		if category != "" {
			filter.Category = &category
		}
		if cmd.Flags().Changed("min-price") {
			filter.MinPrice = &minPrice
		}
		if cmd.Flags().Changed("max-price") {
			filter.MaxPrice = &maxPrice
		}

		products, err := appStore.List(cmd.Context(), filter)
		if err != nil {
			return err
		}

		if output == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(products)
		}

		printTable(products)
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		product, err := appStore.Get(cmd.Context(), id)
		if err != nil {
			return err
		}

		if cmd.Flags().Changed("name") {
			product.Name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("price") {
			p, _ := cmd.Flags().GetFloat64("price")
			if p < 0 {
				return fmt.Errorf("price cannot be negative")
			}
			product.Price = p
		}
		if cmd.Flags().Changed("quantity") {
			product.Quantity, _ = cmd.Flags().GetInt("quantity")
		}
		if cmd.Flags().Changed("category") {
			product.Category, _ = cmd.Flags().GetString("category")
		}

		if err := appStore.Update(cmd.Context(), id, product); err != nil {
			return err
		}

		fmt.Printf("Product updated successfully: %s\n", id)
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Are you sure you want to delete product %s? [y/N]: ", id)
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Deletion cancelled")
				return nil
			}
		}

		if err := appStore.Delete(cmd.Context(), id); err != nil {
			return err
		}

		fmt.Printf("Product deleted successfully\n")
		return nil
	},
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import products from a JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		var products []domain.Product
		if err := json.Unmarshal(data, &products); err != nil {
			return fmt.Errorf("invalid json format: %w", err)
		}

		if err := appStore.BulkImport(cmd.Context(), products); err != nil {
			return err
		}

		fmt.Printf("Successfully imported %d products\n", len(products))
		return nil
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export products to a JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		category, _ := cmd.Flags().GetString("category")

		filter := domain.ListFilter{}
		if category != "" {
			filter.Category = &category
		}

		products, err := appStore.List(cmd.Context(), filter)
		if err != nil {
			return err
		}

		data, err := json.MarshalIndent(products, "", "  ")
		if err != nil {
			return err
		}

		return os.WriteFile(file, data, 0644)
	},
}

func printTable(products []domain.Product) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tPrice\tQuantity\tCategory")
	for _, p := range products {
		fmt.Fprintf(w, "%s\t%s\t%.2f\t%d\t%s\n", p.ID, p.Name, p.Price, p.Quantity, p.Category)
	}
	w.Flush()
}
