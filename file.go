package gotcha

type File struct {
	ast  *dst.File
	path string
}

func NewFileWithPath(path string) (*File, err) {
	_, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return &File{
		ast:  f,
		path: path,
	}, nil
}

func NewFileWithCode(code string) (*File, err) {
	f, err := decorator.Parse(code)
	if err != nil {
		return err
	}

	return &File{
		ast: f,
	}, nil
}

func (f *File) HasFunc(name string) bool {
	fn := f.getFunc(name)
	return fn != nil
}

func (f *File) MergeCode(code string) error {
	merge, err := NewFileWithCode(code)
	if err != nil {
		return err
	}

	f.mergeConstants(merge)
	f.mergeImports(merge)
	f.mergeFuncDecls(merge)
	f.mergeInterfaceDecls(merge)

	// TODO merge other decs
	return nil
}

func (f *File) Save() error {
	if f.path == "" {
		return fmt.Errorf("path not set for file")
	}

	return f.SaveToFile(f.path)
}

func (f *File) SaveToFile(path string) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}
	fh, err := os.Create(path)
	if err != nil {
		return err
	}

	decorator.Fprint(fh, f.ast)
	// runGoImports(path)
}

func (f *File) mergeFuncDecls(file *File) {
	funcDecls := file.getFuncDecls()
	f.ast.Decls = append(f.ast.Decls, funcDecls...)
}

func (f *File) mergeInterfaceDecls(file *File) {
	fSpecs := f.getInterfaceSpecs()
	mSpecs := file.getInterfaceSpecs()

	for _, mSpec := range mSpecs {
		for _, fSpec := range fSpecs {
			if mSpec.Name.Name == fSpec.Name.Name {
				fs := fSpec.Type.(*dst.InterfaceType)
				ms := mSpec.Type.(*dst.InterfaceType)

				fs.Methods.List = append(fs.Methods.List, ms.Methods.List...)
			}
		}
	}
}

func (f *File) mergeConstants(file *File) {
	constDecl := f.getGenDecl(token.CONST)
	if constDecl == nil {
		constDecl = file.getGenDecl(token.CONST)
		if constDecl == nil {
			return
		}

		f.ast.Decls = append([]dst.Decl{constDecl}, f.ast.Decls...)
		return
	}

	mSpecs := file.getGenSpecs(token.CONST)
	constDecl.Specs = append(constDecl.Specs, mSpecs...)
}

func (f *File) mergeImports(file *File) {
	importDecl := f.getGenDecl(token.IMPORT)
	if importDecl == nil {
		importDecl = file.getGenDecl(token.IMPORT)
		if importDecl == nil {
			return
		}

		f.ast.Decls = append([]dst.Decl{importDecl}, f.ast.Decls...)
		return
	}

	mSpecs := file.getGenSpecs(token.IMPORT)
	importDecl.Specs = append(importDecl.Specs, mSpecs...)

	var stdDecs []dst.Spec
	var form3Decs []dst.Spec
	var otherDecs []dst.Spec

	for _, spec := range importDecl.Specs {
		s := spec.(*dst.ImportSpec)
		if strings.Count(s.Path.Value, "/") < 2 {
			spec.Decorations().Before = dst.None
			spec.Decorations().After = dst.None
			stdDecs = append(stdDecs, spec)
		} else if strings.Contains(s.Path.Value, "github.com/form3tech") {
			spec.Decorations().Before = dst.None
			spec.Decorations().After = dst.None
			form3Decs = append(form3Decs, spec)
		} else {
			spec.Decorations().Before = dst.None
			spec.Decorations().After = dst.None
			otherDecs = append(otherDecs, spec)
		}
	}

	if len(stdDecs) > 0 {
		stdDecs[len(stdDecs)-1].Decorations().After = dst.EmptyLine
	}

	if len(form3Decs) > 0 {
		form3Decs[len(form3Decs)-1].Decorations().After = dst.EmptyLine
	}

	if len(otherDecs) > 0 {
		otherDecs[len(otherDecs)-1].Decorations().After = dst.EmptyLine
	}

	importDecl.Specs = append(stdDecs, form3Decs...)
	importDecl.Specs = append(importDecl.Specs, otherDecs...)

	return
}

func (f *File) getInterfaceSpecs() []*dst.TypeSpec {
	interfaceSpecs := []*dst.TypeSpec{}
	for _, decl := range f.ast.Decls {
		if gen, ok := decl.(*dst.GenDecl); ok && gen.Tok == token.TYPE {
			for _, spec := range gen.Specs {
				s, _ := spec.(*dst.TypeSpec)
				if _, ok := s.Type.(*dst.InterfaceType); ok {
					interfaceSpecs = append(interfaceSpecs, s)
				}
			}
		}
	}

	return interfaceSpecs
}

func (f *File) getFunc(name string) *dst.FuncDecl {
	for _, decl := range f.ast.Decls {
		if fn, ok := decl.(*dst.FuncDecl); ok {
			if fn.Name.Name == name {
				return fn
			}
		}
	}

	return nil
}

func (f *File) getFuncDecls() []dst.Decl {
	decls := []dst.Decl{}
	for _, decl := range f.ast.Decls {
		if _, ok := decl.(*dst.FuncDecl); ok {
			decls = append(decls, decl)
		}
	}

	return decls
}

func (f *File) getFuncDecl(name string) *dst.FuncDecl {
	for _, decl := range f.ast.Decls {
		if f, ok := decl.(*dst.FuncDecl); ok {
			if f.Name.Name == name {
				return f
			}
		}
	}

	return nil
}

func (f *File) getFuncDeclWithNameSubstr(substr string) *dst.FuncDecl {
	for _, decl := range f.ast.Decls {
		if f, ok := decl.(*dst.FuncDecl); ok {
			if strings.Contains(f.Name.Name, substr) {
				return f
			}
		}
	}

	return nil
}

func (f *File) getFuncDeclContainingAllStrs(strs []string) *dst.FuncDecl {
	decls := f.getFuncDecls()
	for _, decl := range decls {
		fDecl := decl.(*dst.FuncDecl)
		if _, ok := getFirstStmtIndexContainingAllStrs(fDecl, strs); ok {
			return fDecl
		}
	}

	return nil
}

func (f *File) getGenDecl(tok token.Token) *dst.GenDecl {
	for _, decl := range f.ast.Decls {
		if gen, ok := decl.(*dst.GenDecl); ok && gen.Tok == tok {
			return gen
		}
	}

	return nil
}

func (f *File) getGenSpecs(tok token.Token) []dst.Spec {
	specs := []dst.Spec{}
	for _, decl := range f.ast.Decls {
		if gen, ok := decl.(*dst.GenDecl); ok && gen.Tok == tok {
			specs = append(specs, gen.Specs...)
		}
	}

	return specs
}
