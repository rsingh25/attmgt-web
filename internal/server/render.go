package server

import (
	"attmgt-web/internal/util"
	"embed"
	"fmt"
	"html/template"
	"net/http"
)

type templateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float32
	Data            map[string]any
	CSRFToken       string
	Flash           string
	IsAuthenticated int
	Warning         string
	Error           string
}

var function = template.FuncMap{}

//go:embed "templates"
var templateFS embed.FS

func (s *Server) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}

	/*

		td.StringMap = make(map[string]string)
		td.IntMap = make(map[string]int)
		td.FloatMap = make(map[string]float32)

		if s.session.Exists(r.Context(), "flash") {
			td.Flash = s.session.PopString(r.Context(), "flash")
		}

		if s.session.Exists(r.Context(), "warning") {
			td.Warning = s.session.PopString(r.Context(), "warning")
		}

		if s.session.Exists(r.Context(), "error") {
			td.Error = s.session.PopString(r.Context(), "error")
		}
	*/

	return td
}

func (s *Server) RenderTemplate(w http.ResponseWriter, r *http.Request, page string, data *templateData, partials ...string) error {
	var t *template.Template
	var err error

	pagefile := page + ".page.gohtml"

	_, cachehit := s.templateCache[pagefile]
	if s.env == "prod" && cachehit {
		t = s.templateCache[pagefile]
	} else {
		filesToParse := []string{"templates/base.layout.gohtml"}

		if len(partials) > 0 {
			partialFileNames := util.Map(partials, func(p string) string {
				return fmt.Sprintf("templates/%s.partial.gohtml", p)
			})
			filesToParse = append(filesToParse, partialFileNames...)
		}

		filesToParse = append(filesToParse, "templates/"+pagefile)

		t, err = template.New(pagefile).Funcs(function).ParseFS(templateFS, filesToParse...)
		if err != nil {
			return err
		}
		s.templateCache[pagefile] = t
	}

	//Now apply tempalte data
	if data == nil {
		data = &templateData{}
	}

	data = s.addDefaultData(data, r)
	return t.Execute(w, data)
}
