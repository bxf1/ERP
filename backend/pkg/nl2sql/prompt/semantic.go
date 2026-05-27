package prompt

import "github.com/bxf1/ERP/backend/pkg/nl2sql/models"

// DefaultSemantics returns the built-in semantic mappings for the ERP domain.
func DefaultSemantics() []models.SemanticMapping {
	return []models.SemanticMapping{
		{
			BusinessTerm: "销售额",
			SQLFragment:  "SUM(so.total_amount)",
			Description:  "销售订单总金额汇总",
			Table:        "sales_orders",
		},
		{
			BusinessTerm: "采购金额",
			SQLFragment:  "SUM(po.total_amount)",
			Description:  "采购订单总金额汇总",
			Table:        "purchase_orders",
		},
		{
			BusinessTerm: "库存数量",
			SQLFragment:  "SUM(inv.quantity)",
			Description:  "当前库存总量",
			Table:        "inventory",
		},
		{
			BusinessTerm: "毛利",
			SQLFragment:  "SUM(so.total_amount) - SUM(po.total_amount)",
			Description:  "销售收入减采购成本",
		},
		{
			BusinessTerm: "客户数",
			SQLFragment:  "COUNT(DISTINCT c.id)",
			Description:  "去重客户数量",
			Table:        "customers",
		},
		{
			BusinessTerm: "供应商数",
			SQLFragment:  "COUNT(DISTINCT s.id)",
			Description:  "去重供应商数量",
			Table:        "suppliers",
		},
		{
			BusinessTerm: "本月",
			SQLFragment:  "date_trunc('month', CURRENT_DATE)",
			Description:  "当前月份起始日期",
		},
		{
			BusinessTerm: "上月",
			SQLFragment:  "date_trunc('month', CURRENT_DATE - INTERVAL '1 month')",
			Description:  "上个月起始日期",
		},
		{
			BusinessTerm: "同比增长",
			SQLFragment:  "(current_value - previous_value) / NULLIF(previous_value, 0) * 100",
			Description:  "同比增长率百分比",
		},
		{
			BusinessTerm: "应收金额",
			SQLFragment:  "SUM(so.total_amount) - SUM(COALESCE(so.paid_amount, 0))",
			Description:  "销售订单未收款金额",
			Table:        "sales_orders",
		},
		{
			BusinessTerm: "应付金额",
			SQLFragment:  "SUM(po.total_amount) - SUM(COALESCE(po.paid_amount, 0))",
			Description:  "采购订单未付款金额",
			Table:        "purchase_orders",
		},
	}
}

// MergeSemantics combines default semantics with custom ones, custom taking precedence.
func MergeSemantics(defaults, customs []models.SemanticMapping) []models.SemanticMapping {
	seen := make(map[string]bool)
	var result []models.SemanticMapping

	for _, s := range customs {
		seen[s.BusinessTerm] = true
		result = append(result, s)
	}
	for _, s := range defaults {
		if !seen[s.BusinessTerm] {
			result = append(result, s)
		}
	}
	return result
}
