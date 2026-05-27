package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/bxf1/ERP/backend/internal/errors"
	"github.com/bxf1/ERP/backend/internal/response"
)

type NL2SQLRequest struct {
	Query string `json:"query" binding:"required"`
}

type QueryResult struct {
	Columns     []string                `json:"columns"`
	Rows        []map[string]interface{} `json:"rows"`
	RowCount    int                     `json:"rowCount"`
	ExecutionMs int64                   `json:"executionMs"`
}

type NL2SQLResponse struct {
	SQL         string      `json:"sql"`
	Result      QueryResult `json:"result"`
	Explanation string      `json:"explanation"`
}

type HistoryItem struct {
	ID        string `json:"id"`
	Query     string `json:"query"`
	SQL       string `json:"sql"`
	Timestamp string `json:"timestamp"`
}

var history []HistoryItem

func NL2SQLQuery(c *gin.Context) {
	var req NL2SQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("invalid request"))
		return
	}

	start := time.Now()
	sql, result := mockQuery(req.Query)
	elapsed := time.Since(start).Milliseconds()

	resp := NL2SQLResponse{
		SQL:         sql,
		Result:      result,
		Explanation: mockExplanation(req.Query, result.RowCount),
	}
	resp.Result.ExecutionMs = elapsed

	history = append(history, HistoryItem{
		ID:        time.Now().Format("20060102150405"),
		Query:     req.Query,
		SQL:       sql,
		Timestamp: time.Now().Format(time.RFC3339),
	})

	response.OK(c, resp)
}

func NL2SQLHistory(c *gin.Context) {
	if history == nil {
		history = []HistoryItem{}
	}
	response.OK(c, history)
}

func mockQuery(query string) (string, QueryResult) {
	q := []rune(query)

	if contains(q, "客户") {
		return "SELECT id, name, phone, email, address, created_at FROM customers ORDER BY id",
			QueryResult{
				Columns: []string{"id", "name", "phone", "email", "address", "created_at"},
				Rows: []map[string]interface{}{
					{"id": 1, "name": "张三", "phone": "13800138001", "email": "zhangsan@example.com", "address": "北京市朝阳区", "created_at": "2025-01-15"},
					{"id": 2, "name": "李四", "phone": "13800138002", "email": "lisi@example.com", "address": "上海市浦东新区", "created_at": "2025-02-20"},
					{"id": 3, "name": "王五", "phone": "13800138003", "email": "wangwu@example.com", "address": "广州市天河区", "created_at": "2025-03-10"},
					{"id": 4, "name": "赵六", "phone": "13800138004", "email": "zhaoliu@example.com", "address": "深圳市南山区", "created_at": "2025-04-05"},
				},
				RowCount: 4,
			}
	}

	if contains(q, "销售") || contains(q, "销售额") {
		return "SELECT DATE_FORMAT(order_date, '%Y-%m') AS month, SUM(amount) AS total_sales FROM orders WHERE order_date >= DATE_SUB(NOW(), INTERVAL 12 MONTH) GROUP BY month ORDER BY month",
			QueryResult{
				Columns: []string{"month", "total_sales"},
				Rows: []map[string]interface{}{
					{"month": "2025-06", "total_sales": 125000},
					{"month": "2025-07", "total_sales": 138000},
					{"month": "2025-08", "total_sales": 142000},
					{"month": "2025-09", "total_sales": 156000},
					{"month": "2025-10", "total_sales": 168000},
					{"month": "2025-11", "total_sales": 180000},
					{"month": "2025-12", "total_sales": 195000},
					{"month": "2026-01", "total_sales": 210000},
					{"month": "2026-02", "total_sales": 188000},
					{"month": "2026-03", "total_sales": 205000},
					{"month": "2026-04", "total_sales": 220000},
					{"month": "2026-05", "total_sales": 235000},
				},
				RowCount: 12,
			}
	}

	if contains(q, "地区") || contains(q, "分组") || contains(q, "订单") {
		return "SELECT region, COUNT(*) AS order_count FROM orders GROUP BY region ORDER BY order_count DESC",
			QueryResult{
				Columns: []string{"region", "order_count"},
				Rows: []map[string]interface{}{
					{"region": "华东", "order_count": 1560},
					{"region": "华南", "order_count": 1230},
					{"region": "华北", "order_count": 980},
					{"region": "西南", "order_count": 750},
					{"region": "华中", "order_count": 620},
					{"region": "东北", "order_count": 480},
					{"region": "西北", "order_count": 320},
				},
				RowCount: 7,
			}
	}

	if contains(q, "库存") {
		return "SELECT product_name, stock_quantity, min_stock, CASE WHEN stock_quantity < min_stock THEN '不足' ELSE '正常' END AS status FROM inventory WHERE stock_quantity < min_stock * 1.2 ORDER BY stock_quantity ASC",
			QueryResult{
				Columns: []string{"product_name", "stock_quantity", "min_stock", "status"},
				Rows: []map[string]interface{}{
					{"product_name": "无线蓝牙耳机", "stock_quantity": 15, "min_stock": 50, "status": "不足"},
					{"product_name": "USB-C 数据线", "stock_quantity": 32, "min_stock": 100, "status": "不足"},
					{"product_name": "移动电源 10000mAh", "stock_quantity": 45, "min_stock": 80, "status": "不足"},
					{"product_name": "机械键盘", "stock_quantity": 55, "min_stock": 60, "status": "不足"},
					{"product_name": "显示器支架", "stock_quantity": 72, "min_stock": 70, "status": "正常"},
				},
				RowCount: 5,
			}
	}

	return "SELECT 1 AS result",
		QueryResult{
			Columns:  []string{"result"},
			Rows:     []map[string]interface{}{{"result": 1}},
			RowCount: 1,
		}
}

func mockExplanation(query string, rowCount int) string {
	if rowCount == 0 {
		return "未找到匹配的数据。"
	}
	return "已根据您的查询执行 SQL，返回了相关结果。您可以在下方切换表格和图表视图。"
}

func contains(runes []rune, keyword string) bool {
	s := string(runes)
	k := []rune(keyword)
	for i := 0; i <= len(runes)-len(k); i++ {
		match := true
		for j := 0; j < len(k); j++ {
			if runes[i+j] != k[j] {
				match = false
				break
			}
		}
		if match {
			_ = s
			return true
		}
	}
	return false
}
