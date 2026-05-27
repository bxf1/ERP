// Package semantic defines the semantic layer that maps business concepts
// (metrics, dimensions) to the underlying data model fields.
package semantic

// Layer holds the semantic mapping definitions.
type Layer struct {
	Metrics    []Metric    `json:"metrics"`
	Dimensions []Dimension `json:"dimensions"`
}

// Metric maps a business metric to its SQL expression.
type Metric struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	SQL         string `json:"sql"`
	Unit        string `json:"unit,omitempty"`
}

// Dimension maps a business dimension to a field path.
type Dimension struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	FieldPath   string `json:"field_path"`
}

// DefaultLayer returns the built-in semantic layer definition.
func DefaultLayer() *Layer {
	return &Layer{
		Metrics: []Metric{
			{
				Name:        "total_revenue",
				Label:       "总营收",
				Description: "所有已完成销售订单的总金额",
				SQL:         "SELECT SUM(total_amount) FROM sales_orders WHERE status = 'completed'",
				Unit:        "元",
			},
			{
				Name:        "order_count",
				Label:       "订单数",
				Description: "销售订单总数",
				SQL:         "SELECT COUNT(*) FROM sales_orders",
				Unit:        "个",
			},
			{
				Name:        "inventory_value",
				Label:       "库存价值",
				Description: "当前库存总价值（数量 × 成本价）",
				SQL:         "SELECT SUM(quantity * cost_price) FROM inventory",
				Unit:        "元",
			},
			{
				Name:        "customer_count",
				Label:       "客户数",
				Description: "活跃客户总数",
				SQL:         "SELECT COUNT(*) FROM customers WHERE is_active = true",
				Unit:        "个",
			},
			{
				Name:        "gross_profit",
				Label:       "毛利润",
				Description: "营收 - 成本",
				SQL:         "SELECT SUM(total_amount - cost_amount) FROM sales_orders WHERE status = 'completed'",
				Unit:        "元",
			},
		},
		Dimensions: []Dimension{
			{
				Name:        "product_category",
				Label:       "产品分类",
				Description: "按产品分类维度分析",
				FieldPath:   "products.category",
			},
			{
				Name:        "sales_region",
				Label:       "销售区域",
				Description: "按销售区域维度分析",
				FieldPath:   "customers.region",
			},
			{
				Name:        "time_period",
				Label:       "时间周期",
				Description: "按日/周/月/季度/年维度分析",
				FieldPath:   "orders.created_at",
			},
			{
				Name:        "sales_person",
				Label:       "销售人员",
				Description: "按销售人员维度分析",
				FieldPath:   "orders.sales_person_id",
			},
		},
	}
}
