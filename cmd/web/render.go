package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type templateData struct {
	StringMap            map[string]string
	IntMap               map[string]int
	FloatMap             map[string]float32
	Data                 map[string]interface{}
	CSRFToken            string
	Flash                string
	Warning              string
	Error                string
	IsAuthenticated      int
	API                  string
	CSSVersion           string
	StripeSecretKey      string
	StripePublishableKey string
}

var functions = template.FuncMap{
	"formatCurrency": formatCurrency,
}

func formatCurrency(n int) string {
	f := float32(n / 100)
	return fmt.Sprintf("R$%.2f", f)
}

//go:embed templates
var templateFS embed.FS

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	td.API = app.config.api
	td.StripeSecretKey = app.config.stripe.secret
	td.StripePublishableKey = app.config.stripe.key

	return td
}

func (app *application) renderTemplate(w http.ResponseWriter, r *http.Request, page string, td *templateData, partials ...string) error {
	var t *template.Template

	var err error
	templateToRender := fmt.Sprintf("templates/%s.page.tmpl", page)

	// Check if we have a template entry in the cache for this page already (or the "home" page specifically).
	// If we do, then we use that. If not, we add the page template to the cache, and then parse all of the
	// partial templates and add them to the cache as well.

	_, templateInMap := app.templateCache[templateToRender]

	if app.config.env == "production" && templateInMap {
		t = app.templateCache[templateToRender]
	} else {
		// Initialize a new template map.
		t, err = app.parseTemplate(partials, page, templateToRender)
		if err != nil {
			app.errorLog.Println(err.Error())
			return err
		}
	}

	// check if the templateData struct is nil. If it is, then initialize a new empty map.
	if td == nil {
		td = &templateData{}
	}

	td = app.addDefaultData(td, r)

	// Execute the template set, passing in any dynamic data.
	err = t.Execute(w, td)
	if err != nil {
		app.errorLog.Println(err.Error())
		return err
	}

	return nil
}

func (app *application) parseTemplate(partial []string, page string, templateToRender string) (*template.Template, error) {
	var t *template.Template
	var err error

	// If we have some partials passed in, then we loop through them, parsing the files
	// as we go. The ParseFiles() method returns a *Template pointer, and an error value.
	// If an error is encountered, we return it to the caller.
	if len(partial) > 0 {
		for i, x := range partial {
			partial[i] = fmt.Sprintf("templates/%s.partial.tmpl", x)
		}
	}

	// If we have parsed any partials, we pass the slice of strings to the ParseFiles()
	// method as a variadic parameter. This means that they are appended to the end of
	// the slice, after the page template.
	if len(partial) > 0 {
		t, err = template.New(fmt.Sprintf("%s.page.tmpl", page)).Funcs(functions).ParseFS(templateFS, "templates/base.layout.tmpl", strings.Join(partial, ","), templateToRender)
	} else {
		t, err = template.New(fmt.Sprintf("%s.page.tmpl", page)).Funcs(functions).ParseFS(templateFS, "templates/base.layout.tmpl", templateToRender)
	}
	if err != nil {
		app.errorLog.Println(err.Error())
		return nil, err
	}

	app.templateCache[templateToRender] = t

	return t, nil
}
