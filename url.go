package main

func unquoteURI(uri string) string {
	if len(uri) > 0 && (uri[0] == '\'' || uri[0] == '"') {
		uri = uri[1:]
	}
	if len(uri) > 0 && (uri[len(uri)-1] == '\'' || uri[len(uri)-1] == '"') {
		uri = uri[:len(uri)-1]
	}
	return uri
}
