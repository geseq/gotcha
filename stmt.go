package gotcha

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
