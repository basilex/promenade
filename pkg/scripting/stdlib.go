package scripting

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	lua "github.com/yuin/gopher-lua"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/deal"
	"github.com/basilex/promenade/internal/contexts/order-mgmt/order"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// StandardLibrary provides Promenade-specific API for LUA scripts
type StandardLibrary struct {
	customerUC customer.ICustomerUseCase
	orderUC    order.IUseCase
	dealUC     deal.IUseCase
	db         *sqlx.DB
	ctx        context.Context
}

// NewStandardLibrary creates a new standard library instance
func NewStandardLibrary(ctx context.Context, customerUC customer.ICustomerUseCase, orderUC order.IUseCase, dealUC deal.IUseCase, db *sqlx.DB) *StandardLibrary {
	return &StandardLibrary{
		customerUC: customerUC,
		orderUC:    orderUC,
		dealUC:     dealUC,
		db:         db,
		ctx:        ctx,
	}
}

// Register registers all standard library modules to LUA state
func (s *StandardLibrary) Register(L *lua.LState) {
	s.registerCustomerModule(L)
	s.registerOrderModule(L)
	s.registerDealModule(L)
	s.registerNotifyModule(L)
	s.registerQueryModule(L)
	s.registerDateModule(L)
}

// registerCustomerModule registers Customer API
func (s *StandardLibrary) registerCustomerModule(L *lua.LState) {
	customerTable := L.NewTable()

	// Customer.GetTier(customerID) -> string
	L.SetField(customerTable, "GetTier", L.NewFunction(func(L *lua.LState) int {
		customerIDStr := L.CheckString(1)
		customerID, err := uuidv7.Parse(customerIDStr)
		if err != nil {
			L.RaiseError("invalid customer ID: %s", err.Error())
			return 0
		}

		cust, err := s.customerUC.GetCustomer(s.ctx, customerID)
		if err != nil {
			L.RaiseError("failed to get customer: %s", err.Error())
			return 0
		}

		L.Push(lua.LString(string(cust.Tier)))
		return 1
	}))

	// Customer.SetTier(customerID, tier)
	L.SetField(customerTable, "SetTier", L.NewFunction(func(L *lua.LState) int {
		customerIDStr := L.CheckString(1)
		tierStr := L.CheckString(2)
		
		customerID, err := uuidv7.Parse(customerIDStr)
		if err != nil {
			L.RaiseError("invalid customer ID: %s", err.Error())
			return 0
		}

		// Parse tier
		tier := customer.CustomerTier(tierStr)
		
		err = s.customerUC.UpgradeCustomerTier(s.ctx, customerID, tier)
		if err != nil {
			L.RaiseError("failed to set customer tier: %s", err.Error())
			return 0
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	// Customer.GetStatus(customerID) -> string
	L.SetField(customerTable, "GetStatus", L.NewFunction(func(L *lua.LState) int {
		customerIDStr := L.CheckString(1)
		customerID, err := uuidv7.Parse(customerIDStr)
		if err != nil {
			L.RaiseError("invalid customer ID: %s", err.Error())
			return 0
		}

		cust, err := s.customerUC.GetCustomer(s.ctx, customerID)
		if err != nil {
			L.RaiseError("failed to get customer: %s", err.Error())
			return 0
		}

		L.Push(lua.LString(string(cust.Status)))
		return 1
	}))

	L.SetGlobal("Customer", customerTable)
}

// registerOrderModule registers Order API
func (s *StandardLibrary) registerOrderModule(L *lua.LState) {
	orderTable := L.NewTable()

	// Order.GetStatus(orderID) -> string
	L.SetField(orderTable, "GetStatus", L.NewFunction(func(L *lua.LState) int {
		orderIDStr := L.CheckString(1)
		orderID, err := uuidv7.Parse(orderIDStr)
		if err != nil {
			L.RaiseError("invalid order ID: %s", err.Error())
			return 0
		}

		ord, err := s.orderUC.GetOrder(s.ctx, orderID)
		if err != nil {
			L.RaiseError("failed to get order: %s", err.Error())
			return 0
		}

		L.Push(lua.LString(string(ord.Status)))
		return 1
	}))

	// Order.GetTotal(orderID) -> number
	L.SetField(orderTable, "GetTotal", L.NewFunction(func(L *lua.LState) int {
		orderIDStr := L.CheckString(1)
		orderID, err := uuidv7.Parse(orderIDStr)
		if err != nil {
			L.RaiseError("invalid order ID: %s", err.Error())
			return 0
		}

		ord, err := s.orderUC.GetOrder(s.ctx, orderID)
		if err != nil {
			L.RaiseError("failed to get order: %s", err.Error())
			return 0
		}

		// Return total as cents
		L.Push(lua.LNumber(ord.Total.Amount))
		return 1
	}))

	L.SetGlobal("Order", orderTable)
}

// registerDealModule registers Deal API
func (s *StandardLibrary) registerDealModule(L *lua.LState) {
	dealTable := L.NewTable()

	// Deal.Approve(dealID)
	L.SetField(dealTable, "Approve", L.NewFunction(func(L *lua.LState) int {
		dealIDStr := L.CheckString(1)
		dealID, err := uuidv7.Parse(dealIDStr)
		if err != nil {
			L.RaiseError("invalid deal ID: %s", err.Error())
			return 0
		}

		_, err = s.dealUC.MarkDealAsWon(s.ctx, dealID, "Approved via LUA script")
		if err != nil {
			L.RaiseError("failed to approve deal: %s", err.Error())
			return 0
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	// Deal.Reject(dealID, reason)
	L.SetField(dealTable, "Reject", L.NewFunction(func(L *lua.LState) int {
		dealIDStr := L.CheckString(1)
		reason := L.CheckString(2)
		
		dealID, err := uuidv7.Parse(dealIDStr)
		if err != nil {
			L.RaiseError("invalid deal ID: %s", err.Error())
			return 0
		}

		_, err = s.dealUC.MarkDealAsLost(s.ctx, dealID, reason)
		if err != nil {
			L.RaiseError("failed to reject deal: %s", err.Error())
			return 0
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	// Deal.GetStage(dealID) -> string
	L.SetField(dealTable, "GetStage", L.NewFunction(func(L *lua.LState) int {
		dealIDStr := L.CheckString(1)
		dealID, err := uuidv7.Parse(dealIDStr)
		if err != nil {
			L.RaiseError("invalid deal ID: %s", err.Error())
			return 0
		}

		d, err := s.dealUC.GetDeal(s.ctx, dealID)
		if err != nil {
			L.RaiseError("failed to get deal: %s", err.Error())
			return 0
		}

		L.Push(lua.LString(string(d.Stage)))
		return 1
	}))

	L.SetGlobal("Deal", dealTable)
}

// registerNotifyModule registers Notify API
func (s *StandardLibrary) registerNotifyModule(L *lua.LState) {
	notifyTable := L.NewTable()

	// Notify.SendEmail(to, template)
	L.SetField(notifyTable, "SendEmail", L.NewFunction(func(L *lua.LState) int {
		// TODO: Implement actual email sending
		// to := L.CheckString(1)
		// template := L.CheckString(2)
		L.Push(lua.LBool(true))
		return 1
	}))

	// Notify.SendSMS(phone, message)
	L.SetField(notifyTable, "SendSMS", L.NewFunction(func(L *lua.LState) int {
		// TODO: Implement actual SMS sending
		// phone := L.CheckString(1)
		// message := L.CheckString(2)
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("Notify", notifyTable)
}

// registerQueryModule registers Query API (read-only)
func (s *StandardLibrary) registerQueryModule(L *lua.LState) {
	queryTable := L.NewTable()

	// Query.Execute(sql) -> table
	L.SetField(queryTable, "Execute", L.NewFunction(func(L *lua.LState) int {
		sql := L.CheckString(1)

		// Validate it's SELECT only (security check)
		upperSQL := strings.ToUpper(strings.TrimSpace(sql))
		if !strings.HasPrefix(upperSQL, "SELECT") {
			L.RaiseError("only SELECT queries allowed")
			return 0
		}

		// Block dangerous keywords
		dangerousKeywords := []string{"DELETE", "UPDATE", "INSERT", "DROP", "CREATE", "ALTER", "TRUNCATE", "EXEC", "EXECUTE"}
		for _, keyword := range dangerousKeywords {
			if strings.Contains(upperSQL, keyword) {
				L.RaiseError("dangerous keyword detected: %s", keyword)
				return 0
			}
		}

		// Execute query
		rows, err := s.db.QueryContext(s.ctx, sql)
		if err != nil {
			L.RaiseError("query failed: %s", err.Error())
			return 0
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				// Log error but don't raise - defer should not panic
				// In production, this would log to logger
				_ = closeErr
			}
		}()

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			L.RaiseError("failed to get columns: %s", err.Error())
			return 0
		}

		// Build result array
		result := L.NewTable()
		rowIndex := 1

		for rows.Next() {
			// Create values slice for scanning
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			err := rows.Scan(valuePtrs...)
			if err != nil {
				L.RaiseError("failed to scan row: %s", err.Error())
				return 0
			}

			// Create row table
			rowTable := L.NewTable()
			for i, col := range columns {
				val := values[i]
				
				// Convert to LUA value
				var luaVal lua.LValue
				switch v := val.(type) {
				case nil:
					luaVal = lua.LNil
				case int64:
					luaVal = lua.LNumber(v)
				case float64:
					luaVal = lua.LNumber(v)
				case bool:
					luaVal = lua.LBool(v)
				case []byte:
					luaVal = lua.LString(string(v))
				case string:
					luaVal = lua.LString(v)
				default:
					luaVal = lua.LString(fmt.Sprintf("%v", v))
				}
				
				rowTable.RawSetString(col, luaVal)
			}

			result.RawSetInt(rowIndex, rowTable)
			rowIndex++
		}

		if err := rows.Err(); err != nil {
			L.RaiseError("row iteration error: %s", err.Error())
			return 0
		}

		L.Push(result)
		return 1
	}))

	L.SetGlobal("Query", queryTable)
}

// registerDateModule registers Date utilities
func (s *StandardLibrary) registerDateModule(L *lua.LState) {
	dateTable := L.NewTable()

	// Date.Now() -> string (ISO 8601)
	L.SetField(dateTable, "Now", L.NewFunction(func(L *lua.LState) int {
		now := time.Now().Format(time.RFC3339)
		L.Push(lua.LString(now))
		return 1
	}))

	// Date.Format(dateStr, format) -> string
	L.SetField(dateTable, "Format", L.NewFunction(func(L *lua.LState) int {
		dateStr := L.CheckString(1)
		format := L.CheckString(2)

		// Parse date
		t, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			L.RaiseError("invalid date format: %s", err.Error())
			return 0
		}

		// Format date (Go layout format)
		formatted := t.Format(format)
		L.Push(lua.LString(formatted))
		return 1
	}))

	// Date.GetMonth() -> number
	L.SetField(dateTable, "GetMonth", L.NewFunction(func(L *lua.LState) int {
		month := int(time.Now().Month())
		L.Push(lua.LNumber(month))
		return 1
	}))

	// Date.GetYear() -> number
	L.SetField(dateTable, "GetYear", L.NewFunction(func(L *lua.LState) int {
		year := time.Now().Year()
		L.Push(lua.LNumber(year))
		return 1
	}))

	// Date.GetDay() -> number
	L.SetField(dateTable, "GetDay", L.NewFunction(func(L *lua.LState) int {
		day := time.Now().Day()
		L.Push(lua.LNumber(day))
		return 1
	}))

	L.SetGlobal("Date", dateTable)
}
