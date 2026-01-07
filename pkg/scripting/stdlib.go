package scripting

import (
	lua "github.com/yuin/gopher-lua"
)

// StandardLibrary provides Promenade-specific API for LUA scripts
type StandardLibrary struct {
	// Future: Add dependencies (UseCases, repositories) here
}

// NewStandardLibrary creates a new standard library instance
func NewStandardLibrary() *StandardLibrary {
	return &StandardLibrary{}
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
		// TODO: Implement actual logic via Customer UseCase
		// customerID := L.CheckString(1)
		L.Push(lua.LString("free"))
		return 1
	}))

	// Customer.SetTier(customerID, tier)
	L.SetField(customerTable, "SetTier", L.NewFunction(func(L *lua.LState) int {
		// TODO: Implement actual logic via Customer UseCase
		// customerID := L.CheckString(1)
		// tier := L.CheckString(2)
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("Customer", customerTable)
}

// registerOrderModule registers Order API
func (s *StandardLibrary) registerOrderModule(L *lua.LState) {
	orderTable := L.NewTable()

	// Order.GetStatus(orderID) -> string
	L.SetField(orderTable, "GetStatus", L.NewFunction(func(L *lua.LState) int {
		// TODO: Implement actual logic via Order UseCase
		// orderID := L.CheckString(1)
		L.Push(lua.LString("pending"))
		return 1
	}))

	// Order.SetStatus(orderID, status)
	L.SetField(orderTable, "SetStatus", L.NewFunction(func(L *lua.LState) int {
		// TODO: Implement actual logic via Order UseCase
		// orderID := L.CheckString(1)
		// status := L.CheckString(2)
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("Order", orderTable)
}

// registerDealModule registers Deal API
func (s *StandardLibrary) registerDealModule(L *lua.LState) {
	dealTable := L.NewTable()

	// Deal.Approve(dealID)
	L.SetField(dealTable, "Approve", L.NewFunction(func(L *lua.LState) int {
		// TODO: Implement actual logic via Deal UseCase
		// dealID := L.CheckString(1)
		L.Push(lua.LBool(true))
		return 1
	}))

	// Deal.Reject(dealID, reason)
	L.SetField(dealTable, "Reject", L.NewFunction(func(L *lua.LState) int {
		// TODO: Implement actual logic via Deal UseCase
		// dealID := L.CheckString(1)
		// reason := L.CheckString(2)
		L.Push(lua.LBool(true))
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
		// TODO: Implement actual SQL execution with validation
		// sql := L.CheckString(1)
		// 1. Validate it's SELECT only
		// 2. Execute query
		// 3. Return results as LUA table
		result := L.NewTable()
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
		// TODO: Return actual current time
		L.Push(lua.LString("2026-01-07T12:00:00Z"))
		return 1
	}))

	// Date.Format(dateStr, format) -> string
	L.SetField(dateTable, "Format", L.NewFunction(func(L *lua.LState) int {
		// TODO: Implement date formatting
		// dateStr := L.CheckString(1)
		// format := L.CheckString(2)
		L.Push(lua.LString("2026-01-07"))
		return 1
	}))

	// Date.GetMonth() -> number
	L.SetField(dateTable, "GetMonth", L.NewFunction(func(L *lua.LState) int {
		// TODO: Return actual current month
		L.Push(lua.LNumber(1))
		return 1
	}))

	L.SetGlobal("Date", dateTable)
}
