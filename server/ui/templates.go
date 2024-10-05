package ui

import "html/template"

type AppTemplateCache = map[string]*template.Template

// cache of HTML template files
func CreateHTMLTemplateCache() (map[string]*template.Template, error) {
	// initialize cache (map)
	cache := make(map[string]*template.Template)

	//
	// page - home
	//
	// list of all HTML template files involved for the home page
	var templateFilesHome []string = []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/components/navbar.tmpl.html",
		"./ui/html/pages/home.tmpl.html",
	}
	// parse the HTML template files for home
	parsedTemplateHome, err := template.ParseFiles(templateFilesHome...)
	if err != nil {
		return nil, err
	}
	// add to the cache
	cache["home"] = parsedTemplateHome

	return cache, nil
}
