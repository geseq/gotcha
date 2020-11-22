package gotcha

type Func struct {
	decl *dst.FuncDecl
}

func newFuncWithDecl(decl *dst.FuncDecl) *Func {
	return &Func{
		decl: decl,
	}
}

func (fn *Func) GetFirstStatementIndexContainingAllStrings(strs []string) (idx int, ok bool) {
	for i, item := range fn.decl.Body.List {
		if doesStmtContainAllStrs(item, strs) {
			idx = i
			ok = true
			return
		}
	}

	return
}

func (fn *Func) GetLastStatementIndexContainingAllStrings(strs []string) (idx int, ok bool) {
	for i, item := range fn.decl.Body.List {
		if doesStmtContainAllStrs(item, strs) {
			idx = i
			ok = true
		}
	}

	return
}

func (fn *Func) insertAtBodyListIndex(stmts []dst.Stmt, insertIdx int, removeSpace bool) *Func {
	if insertIdx < 0 {
		fn.decl.Body.List = append(fn.decl.Body.List, stmts...)
	} else if insertIdx == 0 {
		fn.decl.Body.List = append(stmts, fn.decl.Body.List...)
	} else if insertIdx >= len(fn.decl.Body.List) {
		insertIdx = len(fn.decl.Body.List)
		if removeSpace {
			fn.decl.Body.List[insertIdx-1].Decorations().After = dst.None
		}
		fn.decl.Body.List = append(fn.decl.Body.List, stmts...)
	} else {
		if removeSpace {
			fn.decl.Body.List[insertIdx-1].Decorations().After = dst.None
		}
		fn.decl.Body.List = append(fn.decl.Body.List[:insertIdx], append(stmts, fn.decl.Body.List[insertIdx:]...)...)
	}

	return fn
}

func (fn *Func) appendStatements(stmts []dst.Stmt) *Func {
	return fn.insertInFuncAtBodyListIndex(stmts, -1, false)
}

func (fn *Func) prependStatements(stmts []dst.Stmt) *Func {
	return fn.insertInFuncAtBodyListIndex(stmts, 0, false)
}

func doesStmtContainAllStrs(stmt dst.Stmt, strs []string) bool {
	buf := new(bytes.Buffer)
	dst.Fprint(buf, stmt, dst.NotNilFilter)
	code := buf.String()
	c := true
	for _, str := range strs {
		if !strings.Contains(code, str) {
			c = false
			break
		}
	}

	return c
}

func doesStmtContainAnyStrs(stmt dst.Stmt, strs []string) bool {
	buf := new(bytes.Buffer)
	dst.Fprint(buf, stmt, dst.NotNilFilter)
	code := buf.String()
	for _, str := range strs {
		if strings.Contains(code, str) {
			return true
		}
	}

	return false
}
